# Instalación

```ps
express
```

```ps
npm i json-server --save-dev
```

# Express

Usamos una serie de modulos:

```js
var createError = require('http-errors');
var express = require('express');
var path = require('path');
var cookieParser = require('cookie-parser');
var logger = require('morgan');
```

Entre ellos se encuentran:

- Express
- Gestión de errores http
- Manejo del filesystem
- Gestión de cookies

Declaramos la app express

```js
var app = express();
```

Configuramos el motor que generará las vistas. El motor que usamos es `jade`, y las vistas están en el directorio `views`

```js
// view engine setup
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'jade');
```

Usamos el loger y el parser de jsons:

```js
app.use(logger('dev'));
app.use(express.json());
```

Codificamos las urls:

```js
app.use(express.urlencoded({ extended: false }));
```

Hacemos que se gestionen las cookies:

```js
app.use(cookieParser());
```

E indicamos desde donde servir contenido estático - desde el directorio `public`:

```js
app.use(express.static(path.join(__dirname, 'public')));
```

Configuramos las rutas que vamos a gestionar:

```js
app.use('/', indexRouter);
app.use('/users', usersRouter);
```

y cualquier cosa que no sea esta, se tratará como un error:

```js
app.use(function(req, res, next) {
  next(createError(404));
});
```

Finalmente añadimos una gestión de errores:

```js
app.use(function(err, req, res, next) {
  // set locals, only providing error in development
  res.locals.message = err.message;
  res.locals.error = req.app.get('env') === 'development' ? err : {};
  res.locals.euge="Peto!!"

  // render the error page
  res.status(err.status || 500);
  res.render('error',{title:'Ha petado'});
});
```

## Jade

Todas las vistas usan la definición de `layout`. Definimos un bloque que llamamos `content` que se construye con un _h1_ y un _p_. En cada uno de ellos usamos la variable `title`:

```jade
extends layout

block content
  h1= title
  p Welcome to #{title}
```

En el `layout`:

```jade
doctype html
html
  head
    title= title
    link(rel='stylesheet', href='/stylesheets/style.css')
  body
    block content
```

Indicamos que el _body_ sea el contenido del bloque `content` y el _title_ el valor de la variable `title`. En el controler:

```js
router.get('/', function(req, res, next) {
  res.render('index', { title: 'Express' });
});
```

Indicamos que se use la vista `index.jade` con el modelo `res.local` y con `{ title: 'Express' }`.

## Depuración

Lanzamos desde la __DEBUG CONSOLE__ el comando *node json-serv_adv* para depurar. Si queremos depurar _app.js_ podemos crear una configuración que lance el script _/bin/www_.

# Json-Server

Creamos un [mock para una api](https://github.com/typicode/json-server).

Podemos lanzar el servidor desde la línea de comandos:

```ps
json-server --watch db.json
```

O podemos usar una aplicación _node_ para configurar un comportamiento más avanzado:

```js
const jsonServer = require('json-server')

const server = jsonServer.create()
const router = jsonServer.router('db.json')

const middlewares = jsonServer.defaults()

// Set default middlewares (logger, static, cors and no-cache)
server.use(middlewares)
```

Podemos definir rutas con un comportamiento especial. Por ejemplo, aquí definimos que `/echo` devuelva los query parameters:

```js
// Add custom routes before JSON Server router
server.get('/echo', (req, res) => {
    res.jsonp(req.query)
})
```

Podemos reescribir la petición. Por ejemplo, una petición a `http://localhost:3000/api/posts/1` se convertirá en `http://localhost:3000/posts/1`:

```js
//Reescribe la url
server.use(jsonServer.rewriter({
    '/api/*': '/$1',
    '/blog/:resource/:id/show': '/:resource/:id'
}))
```

Podemos crear un middleware que manipulara la petición. Aquí, por ejemplo, estamos añadiendo un campo en el payload antes de crear un registro:

```js
// To handle POST, PUT and PATCH you need to use a body-parser
// You can use the one used by JSON Server
server.use(jsonServer.bodyParser)
server.use((req, res, next) => {
    if (req.method === 'POST') {
        //Añadimos automáticamente un campo en todos los posts que creemos
        req.body={...req.body, creado : Date.now()}
    }
    // Continue to JSON Server router
    next()
})
```

Podemos customizar donde se expone el servidor:

```js
// Use default router
server.use(router)
server.listen(3000, () => {
    console.log('JSON en ejecución en el puerto 3000')
})
```