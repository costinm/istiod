
# Docs

https://istio.io/latest/docs/reference/config/proxy_extensions/wasm-plugin/

https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/wasm/v3/wasm.proto#extensions-wasm-v3-vmconfig

# Runtime

- v8 - not linked: wasmtime, wamr, wavm
  - https://github.com/bytecodealliance/wasm-micro-runtime/
    - embedded
    - interpreter, aot, jit
    - multi-thread, socket, tensorflow-lite, 
  - https://wasmtime.dev/
    - rust, cranelift
    - can be embedded in rust, c, go
    - amd64, arm64, riscv
  - https://wavm.github.io/ LLVM to native, C++, amd64/arm64-wip, 64bit
    - threads https://github.com/WebAssembly/threads
    - wasi, exceptions, references
- null - linked into envoy binary ( not clear why...)

"allow_precompiled" - wasm may include native code. Close to a .so file

One VM per worker.

"allowed_capabilities":
- https://github.com/proxy-wasm/spec/tree/master/abi-versions/vNEXT
- https://github.com/WebAssembly/WASI/blob/main/phases/snapshot/docs.md#modules
  - 

- https://github.com/tetratelabs/proxy-wasm-go-sdk
  - examples, go API

- https://github.com/mosn/mosn is golang, XDS, WASM proxy

Others:
- workerd(cloudfare) - Apache
  - uses the browser style of API (Fetch, WebCrypto, etc) !!!
  - https://blog.cloudflare.com/workerd-open-source-workers-runtime/
  - `env.AUTH_SERVICE` - can be configured, 'mesh' style
  - https://developers.cloudflare.com/workers/runtime-apis/web-standards/
  - https://developers.cloudflare.com/workers/runtime-apis/fetch/
  - 
