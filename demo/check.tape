# Autopus-ADK Check Demo — Architecture rule enforcement
# Usage: cd demo && vhs check.tape

Output check.gif

Require auto

Set Shell "zsh"
Set FontSize 14
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Dracula"
Set TypingSpeed 35ms
Set WindowBar Colorful
Set WindowBarSize 40
Set CursorBlink true

# Run architecture check — shows OK/WARN/ERROR per file
Type "auto check 2>&1 | head -60"
Enter
Sleep 6s
