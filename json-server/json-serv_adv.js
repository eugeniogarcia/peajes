const jsonServer = require('json-server')

const server = jsonServer.create()
const router = jsonServer.router('db.json')

const middlewares = jsonServer.defaults()

// Set default middlewares (logger, static, cors and no-cache)
server.use(middlewares)

// Add custom routes before JSON Server router
server.get('/echo', (req, res) => {
    res.jsonp(req.query)
})

//Reescribe la url
server.use(jsonServer.rewriter({
    '/api/*': '/$1',
    '/blog/:resource/:id/show': '/:resource/:id'
}))

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

// Use default router
server.use(router)
server.listen(3000, () => {
    console.log('JSON en ejecución en el puerto 3000')
})
