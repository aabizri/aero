# Package fmtp implements support for Flight Message Transfer Protocol v2.0

## Usage

### Create a client
`client,_ := fmtp.NewClient("my id")`

### Connect & associate with a remote endpoint
`a, err := client.Dial("address","id")`

### Send a message
`err := a.SendOperatorString("hello there")`