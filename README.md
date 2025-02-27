# TCP

Script para comunicar y pasar datos/texto entre dos dispositivos

Podes conversar en tiempo real o pasar PDFs, binarios, documentos de word/excel/powerpoint, etc...

Por defecto las conexiónes no son encriptadas, pero se puede habilitar la encriptación para mayor seguridad y privacidad

~~~bash
tcp                                     # se conecta a la dirección por defecto

tcp -H DIRECCION                        # ... a la dirección especificada

tcp -e                                  # encripta la conexión

tcp -p PUERTO                           # especifica un puerto

tcp -l                                  # escucha en la dir. por defecto

tcp -l -H DIRECCION                     # ... en la dirección especificada

tcp -l -e                               # escucha usando encriptación

cat archivo.csv | tcp -H DIRECCION      # envía un archivo

tcp -l > archivo.csv                    # recibe un archivo

tcp -q                                  # suprime los logs

tcp -b SIZE                             # tamaño del buffer para enviar datos

tcp -u                                  # utiliza UDP en lugar de TCP
~~~
