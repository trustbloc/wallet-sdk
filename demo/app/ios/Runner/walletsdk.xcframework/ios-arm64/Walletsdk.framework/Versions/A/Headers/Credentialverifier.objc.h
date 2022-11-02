// Objective-C API for talking to github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credentialverifier Go package.
//   gobind -lang=objc github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credentialverifier
//
// File is generated by gobind. Do not edit.

#ifndef __Credentialverifier_H__
#define __Credentialverifier_H__

@import Foundation;
#include "ref.h"
#include "Universe.objc.h"

#include "Api.objc.h"

@class CredentialverifierVerifier;
@class CredentialverifierVerifyOpts;

/**
 * Verifier is used to verify credentials.
 */
@interface CredentialverifierVerifier : NSObject <goSeqRefInterface> {
}
@property(strong, readonly) _Nonnull id _ref;

- (nonnull instancetype)initWithRef:(_Nonnull id)ref;
/**
 * NewVerifier returns a new Verifier.
 */
- (nullable instancetype)init:(id<ApiKeyHandleReader> _Nullable)keyHandleReader didResolver:(id<ApiDIDResolver> _Nullable)didResolver credentialReader:(id<ApiCredentialReader> _Nullable)credentialReader crypto:(id<ApiCrypto> _Nullable)crypto;
/**
 * Verify verifies the given credential or presentation. See the VerifyOpts struct for more information.
 */
- (BOOL)verify:(CredentialverifierVerifyOpts* _Nullable)verifyOpts error:(NSError* _Nullable* _Nullable)error;
@end

/**
 * VerifyOpts represents the various options for the Verify method.
Only of these three should be used for a given call to Verify. If multiple options are used, then one of them will
take precedence over the other per the following order: CredentialID, RawCredential, RawPresentation.
 */
@interface CredentialverifierVerifyOpts : NSObject <goSeqRefInterface> {
}
@property(strong, readonly) _Nonnull id _ref;

- (nonnull instancetype)initWithRef:(_Nonnull id)ref;
- (nonnull instancetype)init;
/**
 * CredentialID is the ID of the credential to be verified.
A credential with the given ID must be found in this Verifier's credentialReader.
 */
@property (nonatomic) NSString* _Nonnull credentialID;
/**
 * RawCredential is the raw credential to be verified.
 */
@property (nonatomic) NSData* _Nullable rawCredential;
/**
 * RawPresentation is the raw presentation to be verified.
 */
@property (nonatomic) NSData* _Nullable rawPresentation;
@end

/**
 * NewVerifier returns a new Verifier.
 */
FOUNDATION_EXPORT CredentialverifierVerifier* _Nullable CredentialverifierNewVerifier(id<ApiKeyHandleReader> _Nullable keyHandleReader, id<ApiDIDResolver> _Nullable didResolver, id<ApiCredentialReader> _Nullable credentialReader, id<ApiCrypto> _Nullable crypto);

#endif
