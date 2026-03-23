# Autopus-ADK Hero Demo — Claude Code session simulation
# Shows the plan → go → sync workflow as it appears inside Claude Code
# Usage: cd demo && vhs hero.tape

Output hero.gif

Set Shell "zsh"
Set FontSize 15
Set Width 1200
Set Height 750
Set Padding 20
Set Theme "Dracula"
Set TypingSpeed 30ms
Set WindowBar Colorful
Set WindowBarSize 40
Set CursorBlink true

# Simulate Claude Code prompt
Type "# 🐙 Autopus — inside Claude Code"
Enter
Sleep 1s

# Step 1: Plan
Type "/auto plan 'Add OAuth2 with Google and GitHub providers'"
Enter
Sleep 1s
Type 'echo "🐙 Autopus ─────────────────────────"'
Enter
Sleep 0.3s
Type 'echo ""'
Enter
Type 'echo "  📋 PRD generated (Standard mode, 10 sections)"'
Enter
Type 'echo "  📝 SPEC-AUTH-001 created"'
Enter
Type 'echo "    ├── prd.md          Product Requirements"'
Enter
Type 'echo "    ├── spec.md         EARS requirements (5 P0, 3 P1)"'
Enter
Type 'echo "    ├── plan.md         4 tasks assigned to 3 executors"'
Enter
Type 'echo "    ├── acceptance.md   12 Given-When-Then criteria"'
Enter
Type 'echo "    └── research.md     Risks + alternatives"'
Enter
Type 'echo ""'
Enter
Type 'echo "  Next: /auto go SPEC-AUTH-001"'
Enter
Sleep 3s

# Step 2: Go
Type "/auto go SPEC-AUTH-001 --auto --loop"
Enter
Sleep 1s
Type 'echo "🐙 Pipeline ─────────────────────────────────────────────"'
Enter
Sleep 0.3s
Type 'echo "  ✓ Phase 1:   Planning         planner decomposed 4 tasks"'
Enter
Sleep 0.5s
Type 'echo "  ✓ Phase 1.5: Test Scaffold    8 failing tests created (RED)"'
Enter
Sleep 0.5s
Type 'echo "  ✓ Phase 2:   Implementation   3 executors in parallel worktrees"'
Enter
Sleep 0.5s
Type 'echo "  ✓ Phase 2.5: Annotation       @AX tags applied to 6 files"'
Enter
Sleep 0.5s
Type 'echo "  ✓ Phase 3:   Testing          coverage: 58% → 89%"'
Enter
Sleep 0.5s
Type 'echo "  ✓ Phase 4:   Review           TRUST 5: APPROVE | Security: PASS"'
Enter
Sleep 0.5s
Type 'echo "  ───────────────────────────────────────────────────────"'
Enter
Type 'echo "  ✅ 4/4 tasks │ 89% coverage │ 0 security issues │ 3m 47s"'
Enter
Sleep 3s

# Step 3: Sync
Type "/auto sync SPEC-AUTH-001"
Enter
Sleep 1s
Type 'echo "╭────────────────────────────────────╮"'
Enter
Type 'echo "│ 🐙 Pipeline Complete!              │"'
Enter
Type 'echo "│ SPEC-AUTH-001: OAuth2 Auth         │"'
Enter
Type 'echo "│ Tasks: 4/4 │ Coverage: 89%         │"'
Enter
Type 'echo "│ Review: APPROVE                    │"'
Enter
Type 'echo "╰────────────────────────────────────╯"'
Enter
Sleep 4s
