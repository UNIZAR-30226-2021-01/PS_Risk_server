# COMUNICACIÓN CON WEBSOCKETS
Cuando un cliente se conecte a una sala se establecerá una conexión mediante WebSockets. Tanto los parámetros como resultados se envían en formato JSON. Las conexiones son a la URL base https://risk-servidor.herokuapp.com/.

## Establecer una conexión
Cuando un cliente quiera establecer una conexión WebSocket con una sala debe hacerlo en las siguientes URL.

- ## /crearSala
    Cuando un cliente quiera crear una nueva sala debe establecer una conexión a esta URL y enviar el siguiente mensaje.
        
        { 
            "idUsuario": int, 
            "clave": string, 
            "tiempoTurno": int, 
            "nombrePartida": string,
        }
    
    **idUsuario:** identificador numérico del usuario que quiere crear la sala.
    
    **clave:** hash SHA256 de la clave del usuario.

    **tiempoTurno:** tiempo de turno para establecer en la sala.

    **nombrePartida:** nombre para establecer en la sala.
    
- ## /aceptarSala
    Cuando un cliente solicite unirse a una sala ya existente debe establecer una conexión a esta URL y enviar el siguiente mensaje.

        { 
            "idUsuario": int, 
            "clave": string, 
            "idSala": int,
        }
    
    **idUsuario:** identificador numérico del usuario que quiere entrar a la sala.
    
    **clave:** hash SHA256 de la clave del usuario.

    **idSala:** identificador numérico de la sala a la que se quiere unir.

- ## /entrarPartida
    Cuando un cliente solicite unirse a una partida empezada debe establecer una conexión a esta URL y enviar el siguiente mensaje.

        { 
            "idUsuario": int, 
            "clave": string, 
            "idSala": int,
        }
    
    **idUsuario:** identificador numérico del usuario que quiere entrar a la sala.
    
    **clave:** hash SHA256 de la clave del usuario.

    **idSala:** identificador numérico de la partida a la que se quiere unir.

## Mensajes
Durante el transcurso de la conexión tanto el servidor como los clientes se pueden enviar mensajes en formato JSON. Los mensajes estan detallados a continuación.

- ## Errores
    Si ocurre algún error durante el transcurso de la conexión se envía un error en formato JSON.

        {
            "_tipoMensaje":"e",
            "code": int,
            "err": string,
        }

    **_tipoMensaje:** se utiliza para ayudar a la decodificación por parte de los clientes. Para los errores su valor es "e".

    **code:** Indica el tipo de error y la acción a tomar por el cliente. Puede tomar los siguientes valores:
    - 1: Ha ocurrido un error y la operación solicitada no se ha llevado a cabo, no se requiere ninguna acción.
    - 2: Ha ocurrido un error verificando al usuario y la operación solicitada no se ha llevado a cabo, se requiere cerrar la sesión del usuario.
    - 3: Ha ocurrido un error en una la sala de espera, se requiere cerrar la conexión con la sala.

    **err:** Explica qué error ha ocurrido.

- ## Salas de espera
    Estos mensajes se envían para comunicar eventos en las salas de espera.

    - ## Información de sala
        El servidor puede enviar a los distintos participantes de una sala el siguiente mensaje cuando necesite comunicar cambios en la información de la sala.
        
            { 
                "_tipoMensaje":"d", 
                "tiempoTurno": int, 
                "nombrePartida": string, 
                "idSala": int
                "jugadores": [ { "id":int, "nombre":string, "icono":int, "aspecto": int, "sigueVivo":bool, "refuerzos":int, } ] 
            }

        **_tipoMensaje:** se utiliza para ayudar a la decodificación por parte de los clientes. Para este mensaje su valor es "d".

        **tiempoTurno:** tiempo de turno configurado en la sala.

        **nombrePartida:** nombre de la partida que comenzará con los jugadores presentes en la sala.

        **idSala:** identificador numérico de la sala.

        **jugadores:** lista de jugadores que se encuentran en la sala.
        - **id:** identificador numérico del jugador.
        - **nombre:** nombre del jugador.
        - **icono:** icono del jugador.
        - **aspecto:** aspecto del jugador.

    - ## Invitar un participante a la sala
        El creador de la sala puede solicitar que se invite a un usuario a la sala. Debe enviar el siguiente mensaje.
            
            { 
                "idInvitado": int, 
                "tipo":"Invitar", 
            } 

        **idInvitado:** indentificador numérico del usuario a invitar a la sala.
        
        **tipo:** se utiliza para ayudar a la decodificación por parte del servidor. Para este mensaje su valor es "Invitar"

    - ## Iniciar partida
        El creador de la sala puede solicitar el inicio de la partida. Debe enviar el siguiente mensaje.

            {
                "tipo":"Iniciar",
            }

        **tipo:** se utiliza para ayudar a la decodificación por parte del servidor. Para este mensaje su valor es "Iniciar"

- ## Partida
    - ## Territorios 
        Cuando el servidor quiera mandar la información sobre algún territorio la codificará de la siguiente manera.

            {
                "id":int, 
                "jugador":int, 
                "tropas":int,
            }

        - **id:** identificador numérico del territorio.
        - **jugador:** identificador numérico dentro de la partida del jugador propietario del territorio.
        - **tropas:** número de tropas en el territorio.

    - ## Información de partida
        El servidor envía este mensaje a todos los jugadores conectados cuando la partida se inicia y cuando se pasa de turno durante una partida. Cuando un jugador se conecta a una partida también recibe este mensaje.

            {
                "_tipoMensaje": "p",
                "tiempoTurno": int,
                "nombrePartida": string,
                "turnoActual": int,
                "turnoJugador": int,
                "fase": int,
                "ultimoTurno": ISO8601,
                "jugadores": [ { "id":int, "nombre":string, "icono":int, "aspecto":int, "sigueVivo":bool, "refuerzos":int, } ],
                "territorios": [ { "id":int, "jugador":int, "tropas":int } ],
            }

        **_tipoMensaje:** se utiliza para ayudar a la decodificación por parte de los clientes. Para este mensaje su valor es "p".

        **tiempoTurno:** tiempo de turno configurado en la sala.

        **nombrePartida:** nombre de la partida.

        **turnoActual:** número de turno.

        **turnoJugador:** identificador numérico del jugador al que le toca el turno.

        **fase:** número de fase dentro de un turno.

        **ultimoTurno:** fecha del último inicio de turno.

        **jugadores:** lista de jugadores de la partida.
        - **id:** identificador numérico del jugador.
        - **nombre:** nombre del jugador.
        - **icono:** icono del jugador.
        - **aspecto:** aspecto del jugador.
        - **sigueVivo:** indica si el jugador sigue vivo.
        - **refuerzos:** número de tropas de refuerzos de las que dispone el jugador.

        **territorios:** lista de territorios de la partida. 

    - ## Cambio de fase
        Cuando un cliente quiera pasar de fase debe enviar el siguiente mensaje.

            {
                "tipo":"Fase",
            }
        
        **tipo:** se utiliza para ayudar a la decodificación por parte del servidor. Para este mensaje su valor es "Fase".
        
    - ## Confirmación de cambio de fase
         El servidor envía este mensaje a todos los jugadores conectados cuando se cambia de fase en la partida.

            {
                "_tipoMensaje": "f"
            }

        **_tipoMensaje:** se utiliza para ayudar a la decodificación por parte de los clientes. Para este mensaje su valor es "f".

    - ## Refuerzos
        Cuando un cliente quiera reforzar un territorio debe enviar el siguiente mensaje.

            {
                "tipo":"Refuerzos",
                "id":int,
                "tropas":int,
            }

        **tipo:** se utiliza para ayudar a la decodificación por parte del servidor. Para este mensaje su valor es "Refuerzos".

        **id:** id del territorio a reforzar.

        **tropas:** número de refuerzos.

    - ## Confirmación de refuerzos
        El servidor envía este mensaje a todos los jugadores conectados cuando se ha reforzado un territorio.

            {
                "_tipoMensaje":"r",
                "territorio": { "id":int, "jugador":int, "tropas":int }
            }

        **_tipoMensaje:** se utiliza para ayudar a la decodificación por parte de los clientes. Para este mensaje su valor es "r".

        **territorio:** información del territorio reforzado.

    - ## Atacar
        Cuando un cliente quiera reforzar un territorio debe enviar el siguiente mensaje.

            {
                "tipo":"Ataque",
                "origen":int,
                "destino":int,
                "tropas":int,
            }
        
        **tipo:** se utiliza para ayudar a la decodificación por parte del servidor. Para este mensaje su valor es "Ataque".

        **origen:** id del territorio desde donde se ataca.

        **destino:** id del territorio a atacar.

        **tropas:** número de tropas para atacar.

    - ## Confirmación de ataque
        El servidor envía este mensaje a todos los jugadores conectados cuando se ha realizado un ataque.

            {
                "_tipoMensaje":"a",
                "territorioOrigen": { "id":int, "jugador":int, "tropas":int },
                "territorioDestino": { "id":int, "jugador":int, "tropas":int },
                "dadosOrigen": [ int ],
                "dadosDestino": [ int ],
            }
        
        **_tipoMensaje:** se utiliza para ayudar a la decodificación por parte de los clientes. Para este mensaje su valor es "a".

        **territorioOrigen:** información del territorio desde el que se ha atacado.

        **territorioDestino:** información del territorio atacado.
        
        **dadosOrigen:** resultado de los dados del atacante.

        **dadosDestino:** resultado de los dados de la defensa.

    - ## Mover
        Cuando un cliente quiera mover tropas entre sus territorios debe enviar el siguiente mensaje.

            {
                "tipo":"Movimiento",
                "origen":int,
                "destino":int,
                "tropas":int,
            }

        **tipo:** se utiliza para ayudar a la decodificación por parte del servidor. Para este mensaje su valor es "Movimiento".

        **origen:** id del territorio desde donde se mueven las tropas.

        **destino:** id del territorio a donde se mueven las tropas.

        **tropas:** número de tropas que se mueven.

    - ## Confirmación movimiento
        El servidor envía este mensaje a todos los jugadores conectados cuando se ha realizado un movimiento de tropas.

            {
                "_tipoMensaje":"m",
                "territorioOrigen": { "id":int, "jugador":int, "tropas":int },
                "territorioDestino": { "id":int, "jugador":int, "tropas":int },
            }

        **_tipoMensaje:** se utiliza para ayudar a la decodificación por parte de los clientes. Para este mensaje su valor es "a".

        **territorioOrigen:** información del territorio desde el que se han movido las tropas.

        **territorioDestino:** información del territorio al que se han movido las tropas.

    - ## Fin de partida
        El servidor envía este mensaje a todos los jugadores conectados cuando termina el turno del último jugador en pie.

            {
                "_tipoMensaje":"t",
                "ganador": { "id":int, "nombre":string, "icono":int, "aspecto":int, "sigueVivo":bool, "refuerzos":int, },
                "riskos":int,
            }
        
        **_tipoMensaje:** se utiliza para ayudar a la decodificación por parte de los clientes. Para este mensaje su valor es "t".

        **ganador:** datos del ganador de la partida.

        **riskos:** cuántos riskos consigue el ganador de la partida. 
        
