```
		concept:

		serial := hart.Open("COM1")
		defer serial.Close()
		master := hart.NewMaster(serial)
		device := device.UniversalDevice{}

		cmd0 := device.Command0()
		cmd1 := device.Command1()
		cmd2 := device.Command2()
		
		executor.Execute(cmd0)
		executor.Execute(cmd1)
		executor.Execute(cmd2)
```

```	
		SendFrame example:

		com, err := hart.Open("COM1")
		if err != nil {
			log.Fatal(err)
		}

		defer com.Close()

		f0 := hart.FrameZero
		tx := f0.Buffer()
		log.Printf("Tx: %x", tx)

		rx := make([]byte, 128)
		n, err := com.SendFrame(tx, rx)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("RxBuf %v: %02x", n, rx[:n])

		if reply, ok := hart.Parse(rx); ok {
			log.Printf("Reply: %v", *reply)
		}
	*/
```