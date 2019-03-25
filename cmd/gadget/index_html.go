package main

// "Default" index html, stored as go-code
func index_html() string {
	return `<!doctype html>
<!--
Copyright 2018 The Go Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
-->
<html>

<head>
    <meta charset="utf-8">
    <title>Go wasm</title>
</head>

<body>
    <style>
    .red {
        color: red;
    }
    .green {
        color: green;
    }
    </style>

    <script src="/wasm_exec.js"></script>

    <script>
        if (!WebAssembly.instantiateStreaming) { // polyfill
            WebAssembly.instantiateStreaming = async (resp, importObject) => {
                const source = await (await resp).arrayBuffer();
                return await WebAssembly.instantiate(source, importObject);
            };
        }

        const go = new Go();
        let mod, inst;
        WebAssembly.instantiateStreaming(fetch("/lib.wasm"), go.importObject).then(async (result) => {
            mod = result.module;
            inst = result.instance;
            await go.run(inst)
        });

    </script>

    <div id="gadget-content"></div>
</body>

</html>`
}
