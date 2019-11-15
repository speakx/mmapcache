module mmapcache

go 1.13

replace environment => ../../environment/src

replace single => ../../single/src

replace svrdemo => ../../svrdemo/src

require (
	github.com/edsrzf/mmap-go v1.0.0
	golang.org/x/sys v0.0.0-20191113165036-4c7a9d0fe056 // indirect
)
