Output demo.gif

Set FontSize 14
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set Padding 20

Type "# porthog - Cross-platform port management CLI"
Enter
Sleep 500ms

Type "porthog version"
Enter
Sleep 1s

Type "# List all listening ports"
Enter
Sleep 500ms

Type "porthog list"
Enter
Sleep 2s

Type "# Find free ports in a range"
Enter
Sleep 500ms

Type "porthog free --range 8000-9000 --count 5"
Enter
Sleep 1.5s

Type "# JSON output for scripting"
Enter
Sleep 500ms

Type "porthog list 135 --json"
Enter
Sleep 2s

Type "# Kill by port (dry run)"
Enter
Sleep 500ms

Type "porthog kill 99999 --dry-run"
Enter
Sleep 1.5s
