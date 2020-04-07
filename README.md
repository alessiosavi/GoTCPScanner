# GoTCPScanner

A simple multithread port scanner

## Usage

The tool take some input parameters

- `host`: the ip/hostname of the target host
- `port`: the port that you want to verify if is open
- `ports`: the range of ports to scan separated by `-`

**_NOTE_**: You can select only one parameter relatead to the port

## Example

### Build

```bash
> git clone https://github.com/alessiosavi/GoTCPScanner.git
> cd GoTCPScanner
> go build
> strip -s GoTCPScanner
```
**_NOTE_**: Windows user can't build the sotware due to the `getUlimitValue` function, that rely on UNIX syscall in order to retrieve the maximum number of open files that the system can handle. You need to remove that function and remove the following piece of code too:

```go
ulimitCurr, _ := getUlimitValue()
if uint64(t.Concurrency) >= ulimitCurr {
    t.Concurrency = int(float64(ulimitCurr) * 0.7)
    fmt.Printf("Provided a thread factor greater than current ulimit size, setting at MAX [%d] requests\n", t.Concurrency)
}
```

### Run

```bash
> ./GoTCPScanner -host localhost -ports 7000-9000 -ports 10000-11000
```
