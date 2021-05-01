# API SERVIDOR RISK
Esta API utiliza peticiones HTTP POST con los parámetros en formato URL-Encoded y las respuestas en formato JSON. Las peticiones se realizan a la URL base https://risk-servidor.herokuapp.com/.

* [Errores y confirmaciones](#errores-y-confirmaciones)
* [Sistema de usuarios](#sistema-de-usuarios)
    * [JSON de datos completos de usuario](#json-de-datos-completos-de-usuario)
    * [/registrar](#registrar)
    * [/iniciarSesion](#iniciarsesion)
    * [/recargarUsuario](#recargarusuario)
    * [/personalizarUsuario](#personalizarUsuario)
    * [/notificaciones](#notificaciones)
    * [/borrarNotificacionTurno](#borrarnotificacionturno)
* [Sistema de Amigos](#sistema-de-amigos)
    * [/amigos](#amigos)
    * [/enviarSolicitudAmistad](#enviarsolicitudamistad)
    * [/gestionAmistad](#gestionamistad)
* [Tienda](#tienda)
    * [/comprar](#comprar)
* [Partidas](#partidas)
    * [/partidas](#partidas-1)
    * [/rechazarPartida](#rechazarpartida)

## Errores y confirmaciones
Si ocurre algún error en la petición se devuelve un error en formato JSON.

    {
        "code": int,
        "err": string,
    }

**code:** Indica el tipo de error y la acción a tomar por el cliente. Puede tomar los siguientes valores:
- 0: No ha ocurrido ningún error y la operación solicitada se ha llevado a cabo, no se requiere ninguna acción. Se utiliza como confirmación de algunas peticiones.
- 1: Ha ocurrido un error y la operación solicitada no se ha llevado a cabo, no se requiere ninguna acción.
- 2: Ha ocurrido un error verificando al usuario y la operación solicitada no se ha llevado a cabo, se requiere cerrar la sesión del usuario.

**err:** Explica qué error ha ocurrido.

## Sistema de usuarios
Peticiones relacionadas con la creación de cuentas, inicio de sesión y personalización de cuentas.

- ## JSON de datos completos de usuario
    Cuando se tengan que enviar los datos completos sobre un usuario se utilizará este JSON. También se incluye la tienda en él. 

        {
            "usuario": {
                "id": int, 
                "nombre": string, 
                "icono": int, 
                "aspecto": int,
                "correo": string, 
                "riskos": int, 
                "recibeCorreos": bool,
            }
            "iconos": [ { "id": int, "precio": int } ],
            "aspectos": [ { "id": int, "precio": int } ],
            "tiendaIconos": [ { "id": int, "precio": int } ],
            "tiendaAspectos": [ { "id": int, "precio":i nt } ],
        }

    **usuario**
    - **id:** Identificador numérico del usuario en la base de datos.
    - **nombre:** Nombre del usuario.
    - **icono:** Identificador numérico del icono que utiliza el usuario.
    - **aspecto:** Identificador numérico del aspecto que utiliza el usuario.
    - **correo:** Correo del usuario.
    - **riskos:** Cantidad de riskos del usuario.
    - **recibeCorreos:** Indica si el usuario desea recibir correos o no.

    **iconos:** Lista de iconos que tiene comprados el usuario.
    - **id:** identificador numérico del icono.
    - **precio:** precio del icono.
    
    **aspectos:** Lista de aspectos que tiene comprados el usuario.
    - **id:** identificador numérico del aspecto.
    - **precio:** precio del aspecto.

    **tiendaIconos:** Lista de iconos en la tienda.
    - **id:** identificador numérico del icono.
    - **precio:** precio del icono.

    **tiendaAspectos:**
    - **id:** identificador numérico del aspecto.
    - **precio:** precio del aspecto.

- ## /registrar
    Se envían los datos necesarios para poder crear una nueva cuenta de usuario. Si la cuenta se crea correctamente se devuelven los datos completos de usuario. Si no se puede crear la cuenta se devuelve un error.

    - **Parámetros:**
        | Nombre        | Tipo   | Descripción                                       |
        |---------------|--------|---------------------------------------------------|
        | nombre        | string | Nombre para el usuario.                           |
        | correo        | string | Correo para el usuario.                           |
        | clave         | string | Hash SHA256 de la clave del usuario.              |
        | recibeCorreos | bool   | Indica si el usuario quiere recibir correos o no. |

    - **Resultado:**

        JSON de datos completos de usuario o JSON de error. 

- ## /iniciarSesion
    Se envian los datos de inicio de sesión de un usuario y se devuelven los datos completos de usuario. Si los datos de inicio de sesión no coinciden con los de ningún usuario se devuelve un error.

    - **Parámetros:**
        | Nombre  | Tipo   | Descripción                         |
        |---------|--------|-------------------------------------|
        | usuario | string | Correo o nombre de usuario.         |
        | clave   | string | Hash SHA256 de la clave del usuario.|

    - **Resultado:**

        JSON de datos completos de usuario o JSON de error.

- ## /recargarUsuario
    Se envía el identificador numérico de un usuario junto con su clave y se devuelven los datos de un usuario que se encuentren en la base de datos o un error. El identificador numérico no cambia una vez se ha creado la cuenta por lo que esta función permite obtener cambios en el usuario que se han realizado desde otras sesiones.

    - **Parámetros:**
        | Nombre    | Tipo   | Descripción                         |
        |-----------|--------|-------------------------------------|
        | idUsuario | int    | Identificador numérico del usuario. |
        | clave     | string | Hash SHA256 de la clave del usuario.|

    - **Resultado:**

        JSON de datos completos de usuario o JSON de error.

- ## /personalizarUsuario
    Se envía el identificador numérico del usuario y su clave junto con el parámetro a modificar de su cuenta y su nuevo valor. Se devuelve un error. Este será error nulo en caso de que todo haya funcionado correctamente.

    - **Parámetros:**
        | Nombre    | Tipo   | Descripción                         |
        |-----------|--------|-------------------------------------|
        | idUsuario | int    | Identificador numérico del usuario. |
        | nuevoDato | void   | Valor nuevo para el parámetro.      |
        | clave     | string | Hash SHA256 de la clave del usuario |
        | tipo      | string | Parámetro a modificar del usuario.  |

        **tipo** puede tomar los siguientes valores:
        - Aspecto
        - Icono
        - Correo
        - Clave
        - Nombre
    
    - **Resultado:**

        JSON de error.

- ## /notificaciones
    Se envía el identificador numérico de un usuario junto con su clave y se devuelve su lista de notificaciones o un error.

    - **Parámetros:**
        | Nombre    | Tipo   | Descripción                         |
        |-----------|--------|-------------------------------------|
        | idUsuario | int    | Identificador numérico del usuario. |
        | clave     | string | Hash SHA256 de la clave del usuario.|

    - **Resultado:**

        JSON con la lista de notificaciones o JSON de error.

            { 
                "notificaciones": [ { "infoExtra":string, "tipo":string, "idEnvio":int } ], 
            }

        **notificaciones**
        - **infoExtra:** Contiene el nombre del usuario o la sala que ha enviado la notificación.
        - **tipo:** Indica si la notificación es una invitación a una sala, una solicitud de amistad o una notificación de turno. Puede tomar los siguientes valores:
            - Peticion de amistad
            - Invitacion
        - **idEnvio:** Identificador numérico de quien ha enviado la notificación.

- ## /borrarNotificacionTurno
    Se envía el identificador numérico de un usuario junto con su clave y el identificador numérico de la partida en la que es su turno y se devuelve un error. Este será error nulo en caso de que todo haya funcionado correctamente.

    - **Parámetros:**
        | Nombre    | Tipo   | Descripción                          |
        |-----------|--------|--------------------------------------|
        | idUsuario | int    | Identificador numérico del usuario.  |
        | clave     | string | Hash SHA256 de la clave del usuario. |
        | idSala    | int    | Identificador numérico de la partida.|

    - **Resultado:**

        JSON de error.

- ## /borrarCuenta
    Se envía el identificador numérico de un usuario junto con su clave y se devuelve un error vacío si la cuenta se ha podido eliminar, o el error sucedido en caso contrario.

    - **Parámetros:**
        | Nombre    | Tipo   | Descripción                         |
        |-----------|--------|-------------------------------------|
        | idUsuario | int    | Identificador numérico del usuario. |
        | clave     | string | Hash SHA256 de la clave del usuario.|

    - **Resultado:**

        JSON de error.
        
## Sistema de Amigos
Peticiones relacionadas con enviar solicitudes de amistad, aceptarlas, rechazarlas y obtener la lista de amigos.

- ## /amigos
    Se envía el identificador numérico de un usuario junto con su clave y se devuelve su lista de amigos o un error.

    - **Parámetros:**
        | Nombre    | Tipo   | Descripción                         |
        |-----------|--------|-------------------------------------|
        | idUsuario | int    | Identificador numérico del usuario. |
        | clave     | string | Hash SHA256 de la clave del usuario.|
    
    - **Resultado:**

        JSON con la lista de amigos o JSON de error.

            { 
                "amigos": [ { "id":int, "nombre":string, "icono":int, "aspecto":int } ],
            }
        
        **amigos:**
        - **id:** identificador numérico del amigo.
        - **nombre:** nombre del amigo.
        - **icono:** icono que utiliza el amigo.
        - **aspecto:** aspecto que utiliza el amigo.

- ## /enviarSolicitudAmistad
    Se envían los datos del usuario emisor y el nombre del receptor y se devuelve un error. Este será error nulo en caso de que todo haya funcionado correctamente.
    
    - **Parámetros:**
        | Nombre      | Tipo   | Descripción                                                 |
        |-------------|--------|-------------------------------------------------------------|
        | idUsuario   | int    | Identificador numérico del usuario.                         |
        | nombreAmigo | string | Nombre del usuario al que se le quiere enviar la solicitud. |
        | clave       | string | Hash SHA256 de la clave del usuario                         |

    - **Resultado:**

        JSON de error.

- ## /gestionAmistad
    Se envían los datos de 2 usuarios y el tipo de gestión que se quiere hacer entre los 2. Se devuelve un error. Este será error nulo en caso de que todo haya funcionado correctamente.

    - **Parámetros:**
        | Nombre    | Tipo   | Descripción                                 |
        |-----------|--------|---------------------------------------------|
        | idUsuario | int    | Identificador numérico del usuario.         |
        | idAmigo   | int    | Identificador numérico del segundo usuario. |
        | clave     | string | Hash SHA256 de la clave del usuario         |
        | decision  | string | Tipo de gestión a realizar.                 |

        **decision** puede tomar los siguientes valores:
        - Rechazar
        - Aceptar
        - Borrar

    - **Resultado:**

        JSON de error.

## Tienda
Peticiones relacionadas con la compra de elementos estéticos del juego.

- ## /comprar
    Se envían los datos de un usuario y del elemento que desea comprar. Se devuelve un error. Este será error nulo en caso de que todo haya funcionado correctamente.

    - **Parámetros:**
        | Nombre    | Tipo   | Descripción                                               |
        |-----------|--------|-----------------------------------------------------------|
        | idUsuario | int    | Identificador numérico del usuario.                       |
        | cosmetico | int    | Identificador numérico del elemento que se desea comprar. |
        | clave     | string | Hash SHA256 de la clave del usuario                       |
        | tipo      | string | Tipo de elemento a comprar.                               |
    
    - **Resultado:**

        JSON de error.

## Partidas
Peticiones relacionadas con las partidas.

- ## /partidas
    Se envían los datos de un usuario. Se devuelve la lista de partidas en las que el usuario está jugando o un error.

    - **Parámetros:**
        | Nombre    | Tipo   | Descripción                         |
        |-----------|--------|-------------------------------------|
        | idUsuario | int    | Identificador numérico del usuario. |
        | clave     | string | Hash SHA256 de la clave del usuario.|

    - **Resultado**

        JSON con la lista de partidas o JSON de error.

            {
                "partidas": [ { 
                        "id": int, 
                        "nombre": string, 
                        "nombreTurno": string,
                        "turnoActual": int,
                        "tiempoTurno": int, 
                        "ultimoTurno": ISO8601, 
                    } ],   
            }

        **partidas:**
        - **id:** identificador numérico de la partida.
        - **nombre:** nombre de la partida.
        - **nombreTurno:** nombre del jugador que tiene el turno ahora.
        - **turnoActual:** turno actual de la partida.
        - **tiempoTurno:** duración de los turnos en minutos.
        - **ultimoTurno:** fecha del último inicio de turno.

- ## /rechazarPartida
    Se envía el identificador numérico de un usuario junto con su clave y el identificador numérico de una partida y se devuelve un error. Este será error nulo en caso de que todo haya funcionado correctamente.

    - **Parámetros:**
        | Nombre    | Tipo   | Descripción                          |
        |-----------|--------|--------------------------------------|
        | idUsuario | int    | Identificador numérico del usuario.  |
        | clave     | string | Hash SHA256 de la clave del usuario. |
        | idSala    | int    | Identificador numérico de la partida.|

    - **Resultado:**

        JSON de error.