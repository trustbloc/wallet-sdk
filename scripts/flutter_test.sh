cd demo/app
flutter pub get
flutter build apk --debug

cd ../../test/integration
INITIATE_ISSUANCE_URL=$(../../build/bin/integration-cli issuance)

echo "${INITIATE_ISSUANCE_URL}"

cd ../../demo/app
adb reverse tcp:8075 tcp:8075
flutter test integration_test --dart-define=INITIATE_ISSUANCE_URL="${INITIATE_ISSUANCE_URL}"