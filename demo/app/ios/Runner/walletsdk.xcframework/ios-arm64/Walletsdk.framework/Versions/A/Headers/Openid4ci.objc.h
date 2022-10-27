// Objective-C API for talking to github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci Go package.
//   gobind -lang=objc github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci
//
// File is generated by gobind. Do not edit.

#ifndef __Openid4ci_H__
#define __Openid4ci_H__

@import Foundation;
#include "ref.h"
#include "Universe.objc.h"

#include "Api.objc.h"

@class Openid4ciInstance;

/**
 * Instance helps with OpenID4CI operations.
 */
@interface Openid4ciInstance : NSObject <goSeqRefInterface> {
}
@property(strong, readonly) _Nonnull id _ref;

- (nonnull instancetype)initWithRef:(_Nonnull id)ref;
/**
 * NewInstance returns a new OpenID4CI Instance.
 */
- (nullable instancetype)init:(NSData* _Nullable)initiateIssuanceRequest format:(NSString* _Nullable)format clientCredentialReader:(id<ApiCredentialReader> _Nullable)clientCredentialReader keyHandleReader:(id<ApiKeyHandleReader> _Nullable)keyHandleReader didResolver:(id<ApiDIDResolver> _Nullable)didResolver;
/**
 * Authorize does something (TODO: Implement).
 */
- (BOOL)authorize:(NSString* _Nullable)preAuthorizedCode authorizationRedirectEndpoint:(NSString* _Nullable)authorizationRedirectEndpoint error:(NSError* _Nullable* _Nullable)error;
/**
 * RequestCredential does something (TODO: Implement).
 */
- (NSData* _Nullable)requestCredential:(NSString* _Nullable)authCode kid:(NSString* _Nullable)kid error:(NSError* _Nullable* _Nullable)error;
@end

/**
 * NewInstance returns a new OpenID4CI Instance.
 */
FOUNDATION_EXPORT Openid4ciInstance* _Nullable Openid4ciNewInstance(NSData* _Nullable initiateIssuanceRequest, NSString* _Nullable format, id<ApiCredentialReader> _Nullable clientCredentialReader, id<ApiKeyHandleReader> _Nullable keyHandleReader, id<ApiDIDResolver> _Nullable didResolver);

#endif