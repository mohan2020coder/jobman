PS C:\Users\KERC> $env:ANDROID_HOME="C:\Users\KERC\AppData\Local\Android\Sdk"
PS C:\Users\KERC> $env:ANDROID_SDK_ROOT="C:\Users\KERC\AppData\Local\Android\Sdk"
PS C:\Users\KERC> $env:Path += ";$env:ANDROID_HOME\platform-tools;$env:ANDROID_HOME\emulator"

PS C:\Users\KERC> emulator -list-avds
Medium_Phone
PS C:\Users\KERC> emulator -avd Medium_Phone


npx expo start
a

---------------------------

npx expo run:android


-------------------
cd backend
go run .\cmd\server\main.go

---------------
cd web
npm  run build
npm  run dev