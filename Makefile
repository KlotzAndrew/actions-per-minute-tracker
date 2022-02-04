release-bins:
	cp ./Release/win64/actions-per-minute-tracker.exe ActionsPerMinuteTracker-win64.exe
	cp ./Release/win32/actions-per-minute-tracker.exe ActionsPerMinuteTracker-win32.exe

new-cert:
	openssl genrsa -out singing_private.key 2048
	chmod 400 singing_private.key
	openssl req -new -x509 -nodes -sha256 -key singing_private.key -out singing_private.crt
	openssl pkcs12 -export -out singing_private.pfx -inkey singing_private.key -in singing_private.crt

sign:
	signtool sign /f scratch/keys/singing_private.pfx /tr http://timestamp.digicert.com ActionsPerMinuteTracker-win64.exe
	signtool sign /f scratch/keys/singing_private.pfx /tr http://timestamp.digicert.com ActionsPerMinuteTracker-win32.exe
