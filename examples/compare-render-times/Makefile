.PHONY: all measure plot clean

all: measure plot

measure:
	go run main.go

plot:
	./generate-charts.sh

clean:
	rm rendering-times.dat
	rm rendering-times-chart.png