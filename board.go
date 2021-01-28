package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-daq/smbus"
)

// Board Structure to hold all details about a MegaIO Board
type Board struct {
	number  uint8
	address uint8

	conn   *smbus.Conn
	atom   int
	inputs int
}

var boards map[uint8]*Board
var boardInit sync.Once

func getBoard(n uint8) (*Board, error) {
	if n < 0 || n > 3 {
		return nil, fmt.Errorf("MegaIO board number must be between 0 and 3")
	}
	boardInit.Do(func() {
		boards = make(map[uint8]*Board)
	})

	brd, ck := boards[n]
	if ck {
		return brd, nil
	}

	brd = &Board{number: n, address: 0x31 + n}
	var ver string
	var err error
	for tries := 0; tries < 10; tries++ {
		ver, err = brd.getVersion()
		if err == nil {
			break
		}
		log.Printf("%d: %v", tries, ver)
		time.Sleep(250 * time.Millisecond)
	}
	if ver == "" {
		return nil, fmt.Errorf("Unable to establish a connection to the MegaIO Board %d", n)

	}
	log.Printf("Connected to MegaIO Board %d. Version %s", n, ver)
	boards[n] = brd
	return brd, nil
}

// megaIOVersion Get the MegaIO hardware and firmware versions
func (brd *Board) getVersion() (verString string, err error) {
	if err = brd.open(); err != nil {
		return
	}
	defer brd.close()
	var ver [4]uint8
	// Major hardware version number - 0x3c
	ver[0], err = brd.readReg(0x3c)
	if err != nil {
		return
	}
	// Minor hardware version number - 0x3d
	ver[1], err = brd.readReg(0x3d)
	if err != nil {
		return
	}
	// Major firmware version number - 0x3e
	ver[2], err = brd.readReg(0x3e)
	if err != nil {
		return
	}
	// Minor firmware verson number - 0x3f
	ver[3], err = brd.readReg(0x3f)
	if err != nil {
		return
	}
	if ver[0]+ver[1] == 0 {
		return "", fmt.Errorf("Unable to get the board version. Is it connected?")
	}
	verString = fmt.Sprintf("Firmware: v %d.%d, Software v %d.%d", ver[0], ver[1], ver[2], ver[3])
	return
}

func (brd *Board) open() error {
	if brd.atom > 0 && brd.conn != nil {
		brd.atom++
		return nil
	}
	conn, err := smbus.Open(1, brd.address)
	if err != nil {
		log.Printf("Error opening address %d: %s", brd.address, err)
		return err
	}
	brd.conn = conn
	brd.atom = 1
	return nil
}

func (brd *Board) close() error {
	if brd.atom >= 1 {
		brd.atom--
		return nil
	}
	brd.conn.Close()
	brd.atom = 0
	brd.conn = nil
	return nil
}

func (brd *Board) readReg(reg uint8) (val uint8, err error) {
	if err := brd.open(); err != nil {
		return 0, err
	}
	defer brd.close()
	return brd.conn.ReadReg(brd.address, reg)
}

func (brd *Board) writeReg(reg uint8, val uint8) (err error) {
	if err := brd.open(); err != nil {
		return err
	}
	defer brd.close()
	return brd.conn.WriteReg(brd.address, reg, val)
}

func (brd *Board) setRelay(n uint8, val bool) bool {
	if val {
		brd.writeReg(0x1, n)
	} else {
		brd.writeReg(0x2, n)
	}
	time.Sleep(100 * time.Microsecond)
	return brd.getRelay(n)
}

func (brd *Board) getRelay(n uint8) bool {
	regVal, err := brd.readReg(0x00)
	if err != nil {
		log.Printf("Unable to read the relay memory state: %s", err)
		return false
	}
	bitMask := uint8(1 << (n - 1))
	return bitMask&regVal == bitMask
}

//#define GPIO_VAL_MEM_ADD				(u8)0x19
//#define GPIO_SET_MEM_ADD				(u8)0x1a
//#define GPIO_CLR_MEM_ADD				(u8)0x1b
//#define GPIO_DIR_MEM_ADD				(u8)0x1c

func (brd *Board) setGPIO(mask uint8, onoff bool) error {
	valDirection, err := brd.readReg(0x1C)
	if err != nil {
		return fmt.Errorf("Failed to get GPIO direction register value: %s", err)
	}
	time.Sleep(10 * time.Millisecond)
	valRising, err := brd.readReg(0x1F)
	if err != nil {
		return fmt.Errorf("Failed to get rising IRQ register value: %s", err)
	}
	time.Sleep(10 * time.Millisecond)
	valFalling, err := brd.readReg(0x20)
	if err != nil {
		return fmt.Errorf("Failed to get falling IRQ register value: %s", err)
	}

	if onoff {
		valRising |= mask
		valFalling |= mask
		valDirection |= mask
		if err = brd.writeReg(0x1C, valDirection); err != nil {
			return fmt.Errorf("Unable to write GPIO direction register: %s", err)
		}
		time.Sleep(10 * time.Millisecond)
	} else {
		valRising &= ^mask
		valFalling &= ^mask
	}
	if err = brd.writeReg(0x1F, valRising); err != nil {
		return fmt.Errorf("Unable to write rising IRQ register: %s", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err = brd.writeReg(0x20, valFalling); err != nil {
		return fmt.Errorf("Unable to write rising IRQ register: %s", err)
	}

	return nil
}

func (brd *Board) checkGpio(mask uint8) bool {
	regVal, err := brd.readReg(0x19)
	if err != nil {
		log.Printf("Unable to read the GPIO state: %s", err)
		return false
	}
	return mask&regVal == mask
}
