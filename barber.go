package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	durmiendo = iota
	observando
	cortando
)

var stateLog = map[int]string{
	0: "durmiendo",
	1: "observando",
	2: "cortando",
}
var wg *sync.WaitGroup

type Cliente struct {
	name string
}

type Barber struct {
	sync.Mutex
	state   int
	cliente *Cliente
}

func (c *Cliente) String() string {
	return fmt.Sprintf("%p", c)[7:]
}

func NuevoBarbero() (b *Barber) {
	return &Barber{
		state: durmiendo,
	}
}

func barber(b *Barber, wr chan *Cliente, principal chan *Cliente) {
	for {
		b.Lock()
		defer b.Unlock()
		b.state = observando
		b.cliente = nil

		// revisando la sala de espera
		fmt.Printf("Revisando Sala de espera: %d\n", len(wr))
		time.Sleep(time.Millisecond * 100)
		select {
		case c := <-wr:
			CortarCabello(c, b)
			b.Unlock()
		default: // La sala de espera está vacía.
			fmt.Printf("barbero durmiendo - %s\n", b.cliente)
			b.state = durmiendo
			b.cliente = nil
			b.Unlock()
			c := <-principal
			b.Lock()
			fmt.Printf("Despertado por %s\n", c)
			CortarCabello(c, b)
			b.Unlock()
		}
	}
}

func CortarCabello(c *Cliente, b *Barber) {
	b.state = cortando
	b.cliente = c
	b.Unlock()
	fmt.Printf("Cortando cabello a %s\n", c)
	time.Sleep(time.Millisecond * 100)
	b.Lock()
	wg.Done()
	b.cliente = nil
}

func cliente(c *Cliente, b *Barber, wr chan<- *Cliente, principal chan<- *Cliente) {

	time.Sleep(time.Millisecond * 50)
	b.Lock()
	fmt.Printf("cliente %s ve a barbero %s a cliente %s| sala_espera: %d\n",
		c, stateLog[b.state], b.cliente, len(wr))
	switch b.state {
	case durmiendo:
		select {
		case principal <- c:
		default:
			select {
			case wr <- c:
			default:
				wg.Done()
			}
		}
	case cortando:
		select {
		case wr <- c:
		default: //sala de espera completa, dejar barberia
			wg.Done()
		}
	case observando:
		panic("Mientras que el barbero esta observando, ningun cliente debe observar al barbero")
	}
	b.Unlock()
}

func main() {
	b := NuevoBarbero()
	SalaEspera := make(chan *Cliente, 5) // 5 sillas
	principal := make(chan *Cliente, 1)  // solo un trabajador
	go barber(b, SalaEspera, principal)

	time.Sleep(time.Millisecond * 100)
	wg = new(sync.WaitGroup)
	n := 10
	wg.Add(10)
	// Generar Clientes
	for i := 0; i < n; i++ {
		time.Sleep(time.Millisecond * 50)
		c := new(Cliente)
		go cliente(c, b, SalaEspera, principal)
	}

	wg.Wait()
	fmt.Println("No hay mas clientes")
}
