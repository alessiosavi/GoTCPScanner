# GoTCPScanner

A simple multithread port scanner

## Usage

The tool take some input parameters

- `host`: the ip/hostname of the target host
- `port`: the port that you want to verify if is open
- `ports`: the range of ports to scan separated by `-`

**_NOTE_**: You can select only one parameter relatead to the port

## Example

```bash
./GoTCPScanner -host localhost -ports 7000-9000 -ports 10000-11000```
