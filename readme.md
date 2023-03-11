# Test
```
go test
```

# Build
```
go build -o go-sync main.go sync.go
```

# Run 
```
./go-sync -s /tmp/dira -d /temp/dirb -w 10
```

# Help

```
./go-sync --help
```

# Paramater
```
-s source folder
-d destination folder
-w worker number, default 10
```