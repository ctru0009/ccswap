Output demo.gif

Require ccswap

Set Shell "bash"
Set FontSize 28
Set FontFamily "JetBrainsMono Nerd Font, FiraCode Nerd Font, Iosevka, monospace"
Set Width 2000
Set Height 600
Set Theme "Catppuccin Mocha"
Set Padding 48
Set BorderRadius 8
Set WindowBar "Colorful"
Set WindowBarSize 32
Set Framerate 30
Set PlaybackSpeed 1.5
Set TypingSpeed 35ms

# Initial prompt
Type "./ccswap list --config=./demo" Sleep 300ms Enter
Sleep 3s

# Clear and switch provider
Type "clear" Sleep 100ms Enter
Sleep 500ms

Type "./ccswap use zai --config=./demo" Sleep 300ms Enter
Sleep 3s

# Check status after switch
Type "clear" Sleep 100ms Enter
Sleep 500ms

Type "./ccswap status --config=./demo" Sleep 300ms Enter
Sleep 3s

# Show list again with new active
Type "clear" Sleep 100ms Enter
Sleep 500ms

Type "./ccswap list --config=./demo" Sleep 300ms Enter
Sleep 3s

# Final - switch back
Type "clear" Sleep 100ms Enter
Sleep 500ms

Type "./ccswap use anthropic --config=./demo" Sleep 300ms Enter
Sleep 3s
