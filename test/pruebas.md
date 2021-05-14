# Casos de prueba

## /registrar (falta que el automático compruebe los campos que no son "usuario")

	- Datos válidos:
		- Sin correo y recibeCorreos=false - ok
		- Con correo y recibeCorreos=false - ok
		- Con correo y recibeCorreos=true - ok

	- Un campo inválido:
		- Nombre vacío - ok
		- Nombre demasiado largo (más de 20 caracteres) - ok
		- Nombre contiene '@' - ok
		- Nombre ya está usado - ok
		- Correo no vacío inválido - ok
		- Correo vacío con recibeCorreos=true - ok
		- Correo ya está usado - ok
		- Clave vacía - ok
		- Clave demasiado larga (más de 64 caracteres, nunca debería ocurrir) - ok
		- RecibeCorreos no se puede parsear a bool (nunca debería ocurrir) - ok

## /iniciarSesion

	- Datos válidos:
		- Con nombre - ok
		- Con correo - ok

	- Un dato inválido:
		- Nombre no coincide con ningún usuario - ok
		- Correo no coincide con ningún usuario - ok
		- Clave incorrecta para el usuario - ok

## /recargarUsuario

	- Datos válidos - ok

	- Un dato inválido:
		- Id no existe - ok
		- Id no se puede parsear a entero - ok
		- Clave incorrecta para el id - ok

## /personalizarUsuario

	- Datos válidos:
		- Modificar aspecto por uno comprado que no sea el actual
		- Modificar icono por uno comprado que no sea el actual
		- Modificar nombre por uno distinto - ok
		- Modificar clave por una distinta - ok
		- Eliminar correo teniendo recibeCorreos=false - ok
		- Añadir correo no teniendo antes - ok
		- Cambiar de correo - ok
		- Modificar recibeCorreos por el contrario teniendo correo - ok

	- Un dato inválido:
		- Id no existe
		- Id no se puede parsear a entero
		- Clave incorrecta para el id
		- Tipo no es correcto
		- Aspecto no comprado
		- Modificar aspecto pero no es un entero
		- Icono no comprado
		- Modificar icono pero no es un entero
		- Nombre vacío
		- Nombre demasiado largo (más de 20 caracteres)
		- Nombre contiene '@'
		- Nombre ya está usado
		- Nueva clave vacía
		- Nueva clave demasiado larga (más de 256 caracteres, nunca debería ocurrir)
		- Eliminar correo teniendo recibeCorreos=true
		- Correo no vacío inválido
		- Correo ya está usado
		- Marcar recibeCorreos=true no teniendo correo
		- Modificar recibeCorreos pero el valor no se puede parsear a bool

## /notificaciones

	- Datos válidos:
		- Comprobar cuando no hay ninguna - ok
		- Comprobar cuando hay una solicitud de amistad - ok
		- Comprobar cuando hay una invitacion a partida - No se hace con test automaticos
		- Comprobar cuando hay una notificación de turno - No se hace con test automaticos
		- Comprobar cuando hay dos de cada - No se hace con test automaticos

	- Datos inválidos:
		- Id no existe
		- Id no se puede parsear a entero
		- Clave incorrecta para el id

## /borrarNotificacionTurno - No se hace con test automaticos

	- Datos válidos

	- Un dato inválido:
		- Id de usuario no existe
		- Id de usuario no se puede parsear a entero
		- Clave incorrecta para el id de usuario
		- Id de partida no existe
		- Id de partida no se puede parsear a entero
		- Usuario no está en la partida idSala
		- No existe la notificación

## /borrarCuenta

	- Datos válidos - ok

	- Un dato inválido:
		- Id no existe
		- Id no se puede parsear a entero
		- Clave incorrecta para el id

## /olvidoClave - No se hace con test automaticos, tienes que mirar el correo para comprobar si ha funcionado bien

	- Datos válidos

	- Un dato inválido:
		- Correo vacío
		- Correo no coincide con el de ningún usuario

## / restablecerClave - No se hace con test automaticos, tienes que mirar el correo para comprobar si ha funcionado bien y necesita usar la web

	- Datos válidos

	- Un dato inválido:
		- Token no existe
		- Clave vacía
		- Clave demasiado larga (más de 256 caracteres, no debería poder ocurrir)

## /amigos

	- Datos válidos:
		- Comprobar la lista de amigos con un amigo - ok
		- Comprobar con un usuario sin amigos - ok
		- Comprobar con un usuario con varios amigos - No creo que sea necesario

	- Un dato inválido:
		- Id no existe
		- Id no se puede parsear a entero
		- Clave incorrecta para el id

## /enviarSolicitudAmistad

	- Datos válidos:
		- El otro usuario no ha enviado solicitud primero - ok
		- El otro usuario sí ha enviado solicitud primero - ok

	- Un dato inválido:
		- Id no existe
		- Id no se puede parsear a entero
		- Clave incorrecta para el id
		- Nombre no existe
		- Nombre es el del propio usuario
		- Nombre es un usuario que ya es amigo

## /gestionAmistad

	- Datos válidos:
		- Rechazar solicitud - ok
		- Aceptar solicitud - ok
		- Eliminar amigo - ok

	- Un dato inválido:
		- IdUsuario no existe
		- IdUsuario no se puede parsear a entero
		- IdAmigo no se puede parsear a entero
		- Clave incorrecta para idUsuario
		- Decision incorrecto
		- Rechazar solicitud que no existe
		- Aceptar solicitud que no existe
		- Borrar alguien que no es amigo

## /comprar

	- Datos válidos:
		- Comprar aspecto
		- Comprar icono

	- Un dato inválido:
		- Id no existe
		- Id no se puede parsear a entero
		- Clave incorrecta para el id
		- Tipo incorrecto
		- Cosmetico no existe para tipo=Aspecto
		- Cosmetico no existe para tipo=Icono
		- Cosmetico no se puede parsear a entero
		- Usuario no tiene riskos suficientes para comprar aspecto
		- Usuario no tiene riskos suficientes para comprar icono

## /partidas - No se hace con test automaticos

	- Datos válidos:
		- Comprobar sin partidas
		- Comprobar con dos partidas

	- Un dato inválido:
		- Id no existe
		- Id no se puede parsear a entero
		- Clave incorrecta para el id

## /rechazarPartida - No se hace con test automaticos

	- Datos válidos

	- Un dato inválido:
		- Id de usuario no existe
		- Id de usuario no se puede parsear a entero
		- Clave incorrecta para el id de usuario
		- Id de partida no se puede parsear a entero
		- Usuario no está invitado a esa partida