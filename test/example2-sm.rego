package prego

# set the next state based on the current state, and the current evaluation result
nextstate = "B" {
  data.prego_state == "A"
  data.example2.myrule
} else = "A" {
  data.example2.myrule
}