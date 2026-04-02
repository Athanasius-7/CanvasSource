# Build

```bash
git clone https://github.com/Athanasius-7/CanvasSource
cd CanvasSource
CGO_ENABLED=1 go build -o CC *.go
```

# Bin

- Contains the binary of CanvasSource for x86_64 Arch.

# Requirements:

- Go Compiler
- Active Canvas Session

## Updates:
- Changed JSON unmarshal lib to bytedance/sonic.
