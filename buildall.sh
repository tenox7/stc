eval "$(go tool dist list | awk -v FS=/ -v BIN=stc '{ print "GOOS=" $1 " GOARCH=" $2 " go build -o out/" BIN "-" $2 "-" gensub(/windows/, "windows.exe", "g", $1) }' )"
