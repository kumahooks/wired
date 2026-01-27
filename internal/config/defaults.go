package config

const DefaultConfig = `title = "wire(d)"
music_library_path = ""
input_char_limit = 256

[notification]
notification_max_width = 44
notification_max_height = 10
notification_shown_max = 3
notification_duration_secs = 4

[colors]
border = "#6f3d49"
text_inactive = "#44262d"
cursor_fg = "#965363"
notification_info = "#539686"
notification_error = "#774a86"
notification_success = "#639653"

[keybinds]
move_left = ["h", "left"]
move_down = ["j", "down"]
move_up = ["k", "up"]
select = ["enter", "l", "right"]
quit = ["ctrl+c", "q"]
cancel = ["ctrl+c", "esc"]
`
