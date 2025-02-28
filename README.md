# LEEG
## Setup
### install go
LEEG uses `go 1.23.6`. Locate the appropriate version for your OS here: https://go.dev/dl/ 

Any OS/processor _should_ work, but the below have been specifically tested:
#### Windows 64 bit
https://go.dev/dl/go1.23.6.windows-amd64.msi
#### MacOS ARM (M1/M2/M3)
https://go.dev/dl/go1.23.6.darwin-arm64.pkg

### pull down the project
`git clone git@github.com:pumpkinheadgiant/leeg.git`


### install air
`air` gives us live-reloading of our application. Our (compilable) changes will immediately be served.

`go install github.com/air-verse/air@latest`

### install templ
`templ` generates HTML fragments, which we'll compose our application with.

`go install github.com/a-h/templ/cmd/templ@latest`

### install boltbrowser
`boltbrowser` is a CLI browser for inspecting the bbolt db file

`go install github.com/br0xen/boltbrowser`

### install delve
`go-delve` allows VS-Code to locally debug go apps

`go install -v github.com/go-delve/delve/cmd/dlv@latest`

### install tailwind
[tailwind v3](https://v3.tailwindcss.com/) is used for CSS layout simplification. A binary monitors the codebase to determine which specific classes are needed, with JIT compilation to the deployed `/public/styles.css` file.

Find the binary appropriate for your OS here: https://github.com/tailwindlabs/tailwindcss/releases/tag/v3.4.17

Once you've downloaded the binary, rename it to be simple `tailwindcss` (or `tailwindcss.exe` for Windows), and move it into the root of the `leeg` project folder.

For non-windows users:
```shell
mv tailwindcss_full_bin_name tailwindcss
chmod +x tailwindcss
``` 

On Mac, in finder, `cntrl-click` open the file in the Browser to allow it be opened. 

## Run
### VSCode
Open the project and ensure the following VSCode extensions are installed. The versions mentioned are the versions available at time of this writing. These or later versions should suffice.
- `Go v0.44.0`
- `templ-vscode v0.0.30`
- `Markdown Preview Enhanced v0.8.15`

We'll run three processes from terminals. There is a `Makefile` that should work for Linux or Mac users, with corresponding commands provided for Windows users.

#### Run tailwind
##### MacOS/Linux
`make css`

##### Windows
`./tailwindcss -i views/css/app.css -o public/styles.css --watch`

This will start the `tailwindcss` application. You should see something like:
```
Rebuilding...

Done in 110ms.
```
As the other processes start and content starts to flow, more updates will be reflected in this window.
#### Run templ
##### MacOS/Linux
`make templ`

##### Windows
`templ generate --watch --proxy=http://localhost:8818`

This starts the `templ` binary, which compiles `foo.templ` files into `foo.go` files.

#### Run air
`air`   

This starts `air`, which will watch for any changes to our `go` source files. As changes occur, the application will recompile and redeploy "hot".

The output will look like:   
```
  __    _   ___  
 / /\  | | | |_) 
/_/--\ |_| |_| \_ v1.61.7, built with Go go1.23.6
```
followed by a listing of monitored resources, followed by application startup logging.

## Use
The application will be served at http://localhost:8818/ if all pieces are properly aligned.

      
##### Special thanks for the Letter 'L' icon:
<a href="https://www.flaticon.com/free-icons/letter-l" title="letter l icons">Letter l icons created by Hight Quality Icons - Flaticon</a>