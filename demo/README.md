# Demo GIFs

Terminal demo GIFs generated with [VHS](https://github.com/charmbracelet/vhs) and optimized with [gifsicle](https://www.lcdf.org/gifsicle/).

## Prerequisites

```bash
brew install vhs gifsicle
```

## Generate GIFs

```bash
cd demo

# Generate all
for tape in *.tape; do vhs "$tape"; done

# Optimize all
for f in *.gif; do gifsicle -O3 --lossy=80 "$f" -o "$f.opt" && mv "$f.opt" "$f"; done
```

## Files

| Tape | Output | Description | README Section |
|------|--------|-------------|----------------|
| `hero.tape` | `hero.gif` | Claude Code session: plan → go → sync | Top ("See It In Action") |
| `workflow.tape` | `workflow.gif` | spec new → status → skills → platforms | "Three Commands to Ship" |
| `doctor.tape` | `doctor.gif` | Health check + CLI detection | "30-Second Install" |
| `check.tape` | `check.gif` | Architecture rule enforcement | standalone |
