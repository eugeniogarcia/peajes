// json server
const jsonServer = require('json-server')

const server = jsonServer.create()
const router = jsonServer.router('db.json')

const middlewares = jsonServer.defaults()

server.use(middlewares)
server.use(router)
server.listen(3000, () => {
    console.log('JSON en ejecución en el puerto 3000');
    console.log('Prueba con http://localhost:3000/posts/1');
})