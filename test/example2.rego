package example2

hello := "universe"

default myrule = false
myrule {
  input.foo == "bar"
}

return[res] {
  data.prego_state == "B"
  myrule == true
  res := hello
}