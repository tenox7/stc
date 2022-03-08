rm -f out/*
eval "$(go tool dist list | awk -v FS=/ -v BIN=stc '!/^(plan9|ios|js|android)/ { print "GOOS=" $1 " GOARCH=" $2 " go build -o out/" BIN "-" $2 "-" gensub(/windows/, "&.exe", "g", $1) }' )"
