/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

async function jsInitSDK(didResolverURI) {
    let promise = new Promise(function(resolve, reject) {
        resolve('jsInitSDK (JS): ' + didResolverURI);
    });
    let result = await promise;
    console.log(result);
}
