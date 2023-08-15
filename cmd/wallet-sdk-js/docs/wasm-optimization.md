### Wasm optimization findings

1. Project is too big to be compiled with tinygo, it crashes during compilation.
2. Most of the project size comes from the Aries framework. If dependencies to aries are stripped, wasm size is reduced from 35MB to 2MB.
3. Go compiler includes code into wasm based on modules usage. if the module is referenced somewhere in code it will be fully included.
4. Aries support a lot of crypto and sign algorithms that wallet-sdk is not used. Or used only in some cases and not needed for all customers.
5. Also, aries supports some of the VDR methods that are not used by wallet-sdk.
6. Removing some algorithms reduces wasm size a lot. For example, removing BBS+ support reduces the size of wasm by 8MB.
7. Modules in Aries related to DID, VDR, Verifieble Credentials, Crypto, and KMS are highly interconnected, including one result including code from all of them. Changing wasm size from 2MB to 27MB. 