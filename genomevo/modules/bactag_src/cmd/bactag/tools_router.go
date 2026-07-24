package main

import (
"bactag/internal/ocode/homBB"
"bactag/internal/ocode/homBlkReorder"
"bactag/internal/ocode/orthBB"
"bactag/internal/ocode/orthJoin1"
"bactag/internal/ocode/orthJoin2"
"bactag/internal/ocode/orthJoin3"
"bactag/internal/ocode/orthoCombine"
"bactag/internal/ocode/orthoParsing"
"bactag/internal/ocode/patching"
"bactag/internal/ocode/patching1"
"bactag/internal/ocode/progBackbonePrep"
"bactag/internal/ocode/revcomp"
"os"
)

// runSubcommand checks if the first argument is an internal tool and runs it.
func runSubcommand() bool {
if len(os.Args) < 2 {
return false
}

cmd := os.Args[1]

// Shift args for the internal commands so they think they are standalone
shiftedArgs := make([]string, len(os.Args)-1)
shiftedArgs[0] = cmd
copy(shiftedArgs[1:], os.Args[2:])

switch cmd {
case "homBB":
os.Args = shiftedArgs
homBB.Main()
return true
case "homBlkReorder":
os.Args = shiftedArgs
homBlkReorder.Main()
return true
case "orthBB":
os.Args = shiftedArgs
orthBB.Main()
return true
case "orthJoin1":
os.Args = shiftedArgs
orthJoin1.Main()
return true
case "orthJoin2":
os.Args = shiftedArgs
orthJoin2.Main()
return true
case "orthJoin3":
os.Args = shiftedArgs
orthJoin3.Main()
return true
case "orthoCombine":
os.Args = shiftedArgs
orthoCombine.Main()
return true
case "orthoParsing":
os.Args = shiftedArgs
orthoParsing.Main()
return true
case "patching":
os.Args = shiftedArgs
patching.Main()
return true
case "patching1":
os.Args = shiftedArgs
patching1.Main()
return true
case "progBackbonePrep":
os.Args = shiftedArgs
progBackbonePrep.Main()
return true
case "revcomp":
os.Args = shiftedArgs
revcomp.Main()
return true
}
return false
}
