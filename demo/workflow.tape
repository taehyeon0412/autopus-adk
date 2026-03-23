# Autopus-ADK Workflow Demo — plan → go → sync
# Usage: cd demo && vhs workflow.tape
# For the "Three Commands to Ship" section

Output workflow.gif

Require auto

Set Shell "zsh"
Set FontSize 16
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Dracula"
Set TypingSpeed 35ms
Set WindowBar Colorful
Set WindowBarSize 40
Set CursorBlink true

# Step 1: Plan — create a SPEC
Type "# Step 1: Plan — describe what you want"
Enter
Sleep 1s
Type "auto spec new WEBHOOK-001 --title 'Webhook delivery with retry'"
Enter
Sleep 3s

# Step 2: Show the SPEC dashboard
Type "# Step 2: Go — 15 agents build it (shown in AI CLI)"
Enter
Sleep 1s
Type "auto status | grep WEBHOOK"
Enter
Sleep 2s

# Step 3: Skills available for the pipeline
Type "# Step 3: 35 skills power the pipeline"
Enter
Sleep 1s
Type "auto skill list | head -12"
Enter
Sleep 3s

# Show platform reach
Type "# Works across all your AI coding CLIs"
Enter
Sleep 1s
Type "auto platform list"
Enter
Sleep 3s
