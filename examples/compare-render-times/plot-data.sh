#!/bin/bash

(cat <<EOF) | gnuplot
  set title "Rendering time vs. image resolution by concurrency method"
  set terminal png size 1500,800 font 'Verdana,12'
  set output "rendering-times-chart.png"
  set grid
  set key box
  set key top right outside
  set xlabel 'Resolution'
  set xtics rotate by -45
  set ylabel 'Rendering time (s)'
  plot 'rendering-times.dat' using 2:xticlabels(1) with lines title 'Sequential', \
       'rendering-times.dat' using 3 with lines title 'Unlimited', \
       'rendering-times.dat' using 4 with lines title '100 routines', \
       'rendering-times.dat' using 5 with lines title 'Buffered'
EOF