
class Config {
  final String didResolverURI;
  final String evaluateIssuanceURL;
  final String evaluatePresentationURL;
  final String attestationURL;
  final String attestationPayload;

  const Config({
    required this.didResolverURI,
    required this.evaluateIssuanceURL,
    required this.evaluatePresentationURL,
    required this.attestationURL,
    required this.attestationPayload,
  });
}