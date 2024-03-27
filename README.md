# TCP

Script para comunicar y pasar datos/texto entre dos dispositivos

Podes conversar en tiempo real o pasar PDFs, binarios, documentos de word/excel/powerpoint, etc...

~~~bash
tcp                                     # se conecta a la dirección por defecto

tcp -H DIRECCION                        # ... a la especificada

tcp -p PUERTO                           # especifica un puerto

tcp -l                                  # escucha en la dir. por defecto

tcp -l -H DIRECCION                     # ... en la especificada

cat archivo.csv | tcp -H DIRECCION      # envía un archivo

tcp -l > archivo.csv                    # recibe un archivo

tcp -q                                  # suprime los logs

tcp -b SIZE                             # tamaño del buffer para enviar datos
~~~
