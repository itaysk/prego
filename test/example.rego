package main

BEGIN[out] {
  clock := time.clock(time.now_ns())
  out := sprintf("Started at %d:%d:%d", [clock[0], clock[1], clock[2]])
}

MAIN[out] {
  input.foo == "bar"
  out := input
}


MAIN[out] {
  out := upper(input.hello)
}

END[out] {
  clock := time.clock(time.now_ns())
  out := sprintf("Finished at %d:%d:%d", [clock[0], clock[1], clock[2]])
}
