# Autopus-ADK Hero Demo — Claude Code session simulation
# Uses Hide/Show to hide setup, runs simulate-claude.sh for realistic UI
# Usage: cd demo && vhs hero.tape

Output hero.gif

Set Shell "bash"
Set FontSize 15
Set Width 1200
Set Height 750
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 0ms
Set WindowBar Colorful
Set WindowBarSize 40
Set CursorBlink false

# Hide all setup
Hide
Type "export PS1=''"
Enter
Type "clear"
Enter
Sleep 300ms
Show

# Run the simulation script
Type "./simulate-claude.sh"
Enter
Sleep 25s
