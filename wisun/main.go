package wisun

import (
	"io"
	"log"
	"os"

	"github.com/goburrow/serial"
)

func main() {
	port, err := serial.Open(&serial.Config{Address: "/dev/ttyUSB0", BaudRate: 115200})
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	if _, err = port.Write([]byte("SKVER\r\n")); err != nil {
		log.Fatal(err)
	}
	if _, err = io.Copy(os.Stdout, port); err != nil {
		log.Fatal(err)
	}
}
