cd test/integration
INITIATE_ISSUANCE_URL=$(../../build/bin/integration-cli issuance)

echo "${INITIATE_ISSUANCE_URL}"

cd ../../demo/app
flutter test integration_test --dart-define=INITIATE_ISSUANCE_URL="${INITIATE_ISSUANCE_URL}"