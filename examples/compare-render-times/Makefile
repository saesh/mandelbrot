.PHONY: all measure plot clean

all: measure plot

measure:
	go run main.go

plot:
	./plot-data.sh

clean:
	rm rendering-times.dat
	rm rendering-times-chart.png
