animation "player_walk" {
  file         = "sheet.png"
  frame_width  = 8
  frame_height = 8
  frame_count  = 4
  mode         = "loop"
  speed        = 100
  transparent  = true
}

animation "player_die" {
  file         = "sheet.png"
  frame_width  = 8
  frame_height = 8
  frame_count  = 4
  mode         = "once"
  speed        = 150
}

animation "player_bounce" {
  file         = "sheet.png"
  frame_width  = 8
  frame_height = 8
  frame_count  = 4
  row          = 0
  mode         = "pingpong"
  speed        = 80
  background   = "#FF0000"
}
