# ebitengine-sketch/005

_examples/widget_demos/textinput with ext/textinput

works on macOS.

It works on js, but the behavior is wrong.

## build wasm

```
env GOOS=js GOARCH=wasm go build -o main.wasm ./_examples/widget_demos/textinput
```
