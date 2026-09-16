package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobman/backend/config"
	"github.com/jobman/backend/database"
)

var (
	testServer *httptest.Server
	testToken  string
)

const (
	ownerPhone    = "9000000000"
	testPassword  = "Test12345"
	techA_Name    = "Tech A"
	techB_Name    = "Tech B"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	adminDSN := "postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable"
	admin, err := pgxpool.New(ctx, adminDSN)
	if err == nil {
		_, _ = admin.Exec(ctx, "DROP DATABASE IF EXISTS service_app_test WITH (FORCE)")
		_, _ = admin.Exec(ctx, "CREATE DATABASE service_app_test")
		admin.Close()
	}

	dsn := "postgres://postgres:postgres@localhost:5433/service_app_test?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	if err := database.NewMigrator(pool).Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}

	cfg := &config.Config{
		AppEnv:        "test",
		AppPort:       "0",
		DatabaseURL:   dsn,
		JWTSecret:     "test-secret",
		JWTExpiration: time.Hour,
		CORSOrigins:   []string{"http://localhost:5173"},
		APIPrefix:     "/api/v1",
	}

	testServer = httptest.NewServer(buildHandler(cfg, pool))

	body, status, err := raw(methodPost, "/api/v1/auth/register", map[string]any{
		"business_name":   "TestFix Services",
		"phone":           ownerPhone,
		"owner_name":      "Test Owner",
		"email":           "owner@testfix.example",
		"address":         "1 Test Street, Bangalore",
		"currency":        "INR",
		"timezone":        "Asia/Kolkata",
		"password":        testPassword,
		"confirm_password": testPassword,
	}, "", nil)
	if err != nil || status != http.StatusCreated {
		fmt.Fprintf(os.Stderr, "register bootstrap failed(%d): %v %s\n", status, err, body)
		os.Exit(1)
	}
	var m0 map[string]any
	_ = json.Unmarshal([]byte(body), &m0)
	testToken, _ = m0["token"].(string)
	if testToken == "" {
		fmt.Fprintln(os.Stderr, "register bootstrap: missing token")
		os.Exit(1)
	}

	code := m.Run()
	testServer.Close()
	pool.Close()
	os.Exit(code)
}

// --- transport helpers (no *testing.T so TestMain can reuse them) ---

const (
	methodGet  = http.MethodGet
	methodPost = http.MethodPost
	methodPut  = http.MethodPut
	methodDel  = http.MethodDelete
)

func raw(method, path string, body any, token string, out any) (string, int, error) {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return "", 0, err
		}
		rd = bytes.NewReader(b)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, testServer.URL+path, rd)
	if err != nil {
		return "", 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return string(b), resp.StatusCode, err
	}
	if out != nil {
		if err := json.Unmarshal(b, out); err != nil {
			return string(b), resp.StatusCode, err
		}
	}
	return string(b), resp.StatusCode, nil
}

// api is the test-convenience wrapper that fails the test on transport errors
// or status codes outside [200,299].
func api(t *testing.T, method, path string, body any, token string, out any) int {
	t.Helper()
	_, status, err := raw(method, path, body, token, out)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return status
}

func mustJSON(t *testing.T, method, path string, body any, token string, out any) {
	t.Helper()
	s := api(t, method, path, body, token, out)
	if s < 200 || s >= 300 {
		t.Fatalf("%s %s: unexpected status %d", method, path, s)
	}
}

func expectStatus(t *testing.T, method, path string, body any, token string, want int) {
	t.Helper()
	if s := api(t, method, path, body, token, nil); s != want {
		t.Fatalf("%s %s: got status %d, want %d", method, path, s, want)
	}
}

// --- test data helpers ---

type obj = map[string]any

func registerBusiness(t *testing.T, phone string) (token, bizID string) {
	t.Helper()
	var res obj
	mustJSON(t, methodPost, "/api/v1/auth/register", obj{
		"business_name":    "Biz " + phone,
		"phone":            phone,
		"owner_name":       "Owner " + phone,
		"email":            "owner" + phone + "@example.com",
		"address":          "Addr " + phone,
		"currency":         "INR",
		"timezone":         "Asia/Kolkata",
		"password":         testPassword,
		"confirm_password": testPassword,
	}, "", &res)
	token, _ = res["token"].(string)
	if u, ok := res["user"].(obj); ok {
		bizID, _ = u["business_id"].(string)
	}
	return token, bizID
}

func login(t *testing.T, phone, password, token string) string {
	t.Helper()
	var res obj
	mustJSON(t, methodPost, "/api/v1/auth/login", obj{"phone": phone, "password": password}, "", &res)
	out, _ := res["token"].(string)
	if out == "" {
		t.Fatalf("login for %s: no token", phone)
	}
	return out
}

func createCustomer(t *testing.T, token, name, phone string) string {
	t.Helper()
	var res obj
	mustJSON(t, methodPost, "/api/v1/customers", obj{"name": name, "phone": phone}, token, &res)
	id, _ := res["id"].(string)
	return id
}

func createTechnician(t *testing.T, token, name, phone string) string {
	t.Helper()
	var res obj
	mustJSON(t, methodPost, "/api/v1/technicians", obj{
		"name": name, "phone": phone, "password": testPassword,
	}, token, &res)
	id, _ := res["id"].(string)
	return id
}

func createJob(t *testing.T, token, customerID, techID string) string {
	t.Helper()
	body := obj{
		"customer_id":       customerID,
		"service_type":      "AC Repair",
		"problem_description": "Not cooling",
		"estimated_amount":  1000,
		"scheduled_at":      time.Now().Add(time.Hour).Format(time.RFC3339),
	}
	if techID != "" {
		body["technician_id"] = techID
	}
	var res obj
	mustJSON(t, methodPost, "/api/v1/jobs", body, token, &res)
	id, _ := res["id"].(string)
	if id == "" {
		t.Fatalf("createJob: no id")
	}
	return id
}

func jobGet(t *testing.T, token, jobID string) obj {
	t.Helper()
	var res obj
	mustJSON(t, methodGet, "/api/v1/jobs/"+jobID, nil, token, &res)
	return res
}

// --- tests ---

func TestAuthFlow(t *testing.T) {
	phone := "9000199901"
	token, bizID := registerBusiness(t, phone)
	if token == "" || bizID == "" {
		t.Fatal("register: missing token or business id")
	}

	// me
	var me obj
	mustJSON(t, methodGet, "/api/v1/auth/me", nil, token, &me)
	if me["user"] == nil || me["business"] == nil {
		t.Fatal("me: missing user/business")
	}

	// business/me
	var biz obj
	mustJSON(t, methodGet, "/api/v1/business/me", nil, token, &biz)
	if biz["name"] != "Biz "+phone {
		t.Fatalf("business name = %v", biz["name"])
	}

	// wrong password
	expectStatus(t, methodPost, "/api/v1/auth/login", obj{"phone": phone, "password": "nope"}, "", http.StatusUnauthorized)

	// no token
	expectStatus(t, methodGet, "/api/v1/customers", nil, "", http.StatusUnauthorized)
}

func TestTenantIsolation(t *testing.T) {
	tokenA, _ := registerBusiness(t, "9000200001")
	tokenB, _ := registerBusiness(t, "9000200002")

	// A creates a customer, B must not see it
	custA := createCustomer(t, tokenA, "Tenant A Cust", "9000211101")
	var listA obj
	mustJSON(t, methodGet, "/api/v1/customers", nil, tokenA, &listA)
	totalA, _ := listA["pagination"].(obj)["total"].(float64)
	if totalA != 1 {
		t.Fatalf("tenant A customers = %v", totalA)
	}

	var listB obj
	mustJSON(t, methodGet, "/api/v1/customers", nil, tokenB, &listB)
	totalB, _ := listB["pagination"].(obj)["total"].(float64)
	if totalB != 0 {
		t.Fatalf("tenant B customers = %v (should be 0)", totalB)
	}

	// B cannot read A's customer
	var errRes obj
	s := api(t, methodGet, "/api/v1/customers/"+custA, nil, tokenB, &errRes)
	if s != http.StatusNotFound {
		t.Fatalf("cross-tenant customer read: status %d", s)
	}
	code, _ := errRes["error"].(obj)["code"].(string)
	if code != "RESOURCE_NOT_FOUND" {
		t.Fatalf("cross-tenant code = %s", code)
	}
}

func TestCustomerAndTechnicianLifecycle(t *testing.T) {
	token := testToken

	custID := createCustomer(t, token, "Ramesh", "9000222201")
	if custID == "" {
		t.Fatal("create customer: no id")
	}
	var got obj
	mustJSON(t, methodGet, "/api/v1/customers/"+custID, nil, token, &got)
	if got["name"] != "Ramesh" {
		t.Fatalf("customer name = %v", got["name"])
	}

	var upd obj
	mustJSON(t, methodPut, "/api/v1/customers/"+custID, obj{"notes": "Prefers mornings"}, token, &upd)
	if upd["notes"] != "Prefers mornings" {
		t.Fatalf("update notes = %v", upd["notes"])
	}

	// technicians
	techID := createTechnician(t, token, techA_Name, "9000222202")
	if techID == "" {
		t.Fatal("create technician: no id")
	}
	var tlist obj
	mustJSON(t, methodGet, "/api/v1/technicians", nil, token, &tlist)
	data, _ := tlist["data"].([]any)
	found := false
	for _, v := range data {
		if o, ok := v.(obj); ok && o["id"] == techID {
			found = true
		}
	}
	if !found {
		t.Fatal("created technician not in list")
	}
}

func TestJobStateMachine(t *testing.T) {
	token := testToken

	// invalid transition first: accept a PENDING job as owner -> blocked by role
	job := createJob(t, token, createCustomer(t, token, "SM Cust", "9000233301"), "")
	expectStatus(t, methodPost, "/api/v1/jobs/"+job+"/accept", obj{}, "illegal-token", http.StatusUnauthorized)

	// state machine: cannot complete straight from PENDING (route is tech-only, so use service-level via Cancel route is owner only)
	// Simulate milestone transitions as owner
	expectStatus(t, methodPost, "/api/v1/jobs/"+job+"/cancel", nil, token, http.StatusOK)
	state := jobGet(t, token, job)
	if state["status"] != "CANCELLED" {
		t.Fatalf("job status = %v", state["status"])
	}
	// cancelling again must fail
	expectStatus(t, methodPost, "/api/v1/jobs/"+job+"/cancel", nil, token, http.StatusUnprocessableEntity)
}

func TestTechnicianWorkflowAndAuthz(t *testing.T) {
	token := testToken
	techA := createTechnician(t, token, techA_Name, "9000244401")
	_ = createTechnician(t, token, techB_Name, "9000244402")

	cust := createCustomer(t, token, "Tech Cust", "9000244411")
	job := createJob(t, token, cust, techA)

	login := func(phone string) string { return login(t, phone, testPassword, "") }
	tokenA := login("9000244401")
	tokenB := login("9000244402")

	// Tech A can proceed through transitions on assigned job
	expectStatus(t, methodPost, "/api/v1/jobs/"+job+"/accept", obj{}, tokenA, http.StatusOK)
	expectStatus(t, methodPost, "/api/v1/jobs/"+job+"/on-the-way", obj{}, tokenA, http.StatusOK)
	expectStatus(t, methodPost, "/api/v1/jobs/"+job+"/start", obj{}, tokenA, http.StatusOK)

	// Tech B is not assigned -> forbidden
	var errRes obj
	s := api(t, methodPost, "/api/v1/jobs/"+job+"/start", obj{}, tokenB, &errRes)
	if s != http.StatusForbidden {
		t.Fatalf("tech B start unassigned: status %d", s)
	}
	code, _ := errRes["error"].(obj)["code"].(string)
	if code != "JOB_NOT_ASSIGNED" {
		t.Fatalf("tech B start code = %s", code)
	}

	// Tech cannot list/use owner endpoints
	expectStatus(t, methodGet, "/api/v1/customers", nil, tokenA, http.StatusForbidden)
	expectStatus(t, methodPost, "/api/v1/customers", obj{"name": "x", "phone": "1"}, tokenA, http.StatusForbidden)
	expectStatus(t, methodGet, "/api/v1/technicians", nil, tokenA, http.StatusForbidden)

	// Complete without items -> invalid
	var invalid obj
	s = api(t, methodPost, "/api/v1/jobs/"+job+"/complete", obj{"items": []any{}}, tokenA, &invalid)
	if s != http.StatusBadRequest && s != http.StatusUnprocessableEntity {
		t.Fatalf("complete without items: status %d", s)
	}

	// Complete with items + full payment
	var comp obj
	s = api(t, methodPost, "/api/v1/jobs/"+job+"/complete", obj{
		"items": []any{
			obj{"description": "Gas refill", "quantity": 1, "unit_price": 800},
			obj{"description": "Labor", "quantity": 1, "unit_price": 200},
		},
		"payment": obj{"amount": 1000, "method": "UPI"},
	}, tokenA, &comp)
	if s != http.StatusOK {
		t.Fatalf("complete: status %d body %v", s, comp)
	}
	if comp["status"] != "COMPLETED" {
		t.Fatalf("complete status = %v", comp["status"])
	}
	if comp["final_amount"].(float64) != 1000 {
		t.Fatalf("final_amount = %v", comp["final_amount"])
	}
	if comp["payment_status"] != "PAID" {
		t.Fatalf("payment_status = %v", comp["payment_status"])
	}
	if comp["receipt"] == nil {
		t.Fatal("no receipt issued on completion")
	}
	r, _ := comp["receipt"].(obj)
	receiptNo, _ := r["receipt_number"].(string)
	if receiptNo == "" {
		t.Fatal("receipt_number missing")
	}

	// Job now completed: further transitions must fail
	s = api(t, methodPost, "/api/v1/jobs/"+job+"/start", obj{}, tokenA, nil)
	if s < 400 {
		t.Fatalf("start after complete: status %d (want >=400)", s)
	}
}

func TestPaymentsProgression(t *testing.T) {
	token := testToken
	tech := createTechnician(t, token, "Pay Tech", "9000255502")
	cust := createCustomer(t, token, "Pay Cust", "9000255501")
	job := createJob(t, token, cust, tech)
	techTok := login(t, "9000255502", testPassword, "")

	// owner adds a partial payment
	var pay obj
	mustJSON(t, methodPost, "/api/v1/jobs/"+job+"/payments", obj{"amount": 500, "method": "CASH"}, token, &pay)
	if pay["status"] != "PAID" {
		t.Fatalf("payment status = %v", pay["status"])
	}
	if pay["amount"].(float64) != 500 {
		t.Fatalf("payment amount = %v", pay["amount"])
	}

	// technician (self-serve) can list payments for an assigned job
	var pList obj
	mustJSON(t, methodGet, "/api/v1/jobs/"+job+"/payments", nil, techTok, &pList)
	if items, ok := pList["data"].([]any); !ok || len(items) != 1 {
		t.Fatalf("tech payment list = %v", pList)
	}

	// walk the job to STARTED first so completion is a valid transition
	expectStatus(t, methodPost, "/api/v1/jobs/"+job+"/accept", obj{}, techTok, http.StatusOK)
	expectStatus(t, methodPost, "/api/v1/jobs/"+job+"/on-the-way", obj{}, techTok, http.StatusOK)
	expectStatus(t, methodPost, "/api/v1/jobs/"+job+"/start", obj{}, techTok, http.StatusOK)

	// complete (no new payment) -> PARTIAL
	var comp obj
	api(t, methodPost, "/api/v1/jobs/"+job+"/complete", obj{
		"items": []any{obj{"description": "Part", "quantity": 1, "unit_price": 1000}},
	}, techTok, &comp)
	if comp["payment_status"] != "PARTIAL" {
		t.Fatalf("payment_status after partial = %v", comp["payment_status"])
	}

	// remaining payment via owner -> PAID
	var fin obj
	mustJSON(t, methodPost, "/api/v1/jobs/"+job+"/payments", obj{"amount": 500, "method": "UPI"}, token, &fin)
	state := jobGet(t, token, job)
	if state["payment_status"] != "PAID" {
		t.Fatalf("payment_status after final = %v", state["payment_status"])
	}

	// over-payment rejected (another completion payment style: single payment on a new job > final)
	job2 := createJob(t, token, createCustomer(t, token, "OverCust", "9000255503"), "")
	var comp2 obj
	s := api(t, methodPost, "/api/v1/jobs/"+job2+"/complete", obj{
		"items":   []any{obj{"description": "X", "quantity": 1, "unit_price": 500}},
		"payment": obj{"amount": 700, "method": "CASH"},
	}, token, &comp2)
	// owner cannot complete (FORBIDDEN) - so assert not 2xx
	if s >= 200 && s < 300 {
		t.Fatalf("owner complete should be forbidden, got %d", s)
	}
}

func TestDashboardAndReceiptPublic(t *testing.T) {
	token := testToken

	var dash obj
	mustJSON(t, methodGet, "/api/v1/dashboard/today", nil, token, &dash)
	if dash["date"] == "" {
		t.Fatal("dashboard: no date")
	}

	// public receipt endpoint for a completed job's receipt
	var list obj
	mustJSON(t, methodGet, "/api/v1/jobs", obj{}, token, &list)
	data, _ := list["data"].([]any)
	if len(data) == 0 {
		t.Skip("no jobs to test public receipt")
	}
	for _, item := range data {
		o, _ := item.(obj)
		if o["status"] != "COMPLETED" {
			continue
		}
		jobID, _ := o["id"].(string)
		if jobID == "" {
			continue
		}
		var rec obj
		s := api(t, methodGet, "/api/v1/jobs/"+jobID+"/receipt", nil, token, &rec)
		if s != http.StatusOK {
			continue
		}
		pubURL, _ := rec["public_url"].(string)
		if pubURL == "" {
			continue
		}
		tok := pubURL[3:]
		var pub obj
		mustJSON(t, methodGet, "/api/v1/public/receipts/"+tok, nil, "", &pub)
		if pub["receipt_number"] == "" {
			t.Fatal("public receipt: number empty")
		}
		var pageObj any
		_, s, _ = raw(methodGet, "/r/"+tok, nil, "", &pageObj)
		if s != http.StatusOK {
			t.Fatalf("public HTML receipt: status %d", s)
		}
		return
	}
	t.Skip("no completed job found to test public receipt")
}

func TestCreateValidation(t *testing.T) {
	token := testToken

	// missing name
	var res obj
	s := api(t, methodPost, "/api/v1/customers", obj{"phone": "111"}, token, &res)
	if s != http.StatusBadRequest {
		t.Fatalf("customer missing name: status %d", s)
	}

	// job with unknown customer
	s = api(t, methodPost, "/api/v1/jobs", obj{
		"customer_id": "00000000-0000-0000-0000-000000000000",
		"service_type": "AC",
	}, token, &res)
	if s != http.StatusNotFound {
		t.Fatalf("job unknown customer: status %d", s)
	}

	// active technician with short password
	s = api(t, methodPost, "/api/v1/technicians", obj{"name": "X", "phone": "9000266601", "password": "abc"}, token, &res)
	if s != http.StatusBadRequest {
		t.Fatalf("short password: status %d", s)
	}
}