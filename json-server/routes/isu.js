var express = require('express');
var batches = require("../isu/migracionISU");

var router = express.Router();

//Seguimiento de la migración
router.get('/ZWS_JOB_MONITOR', function(req, res, next) {
  res.setHeader("Content-Type","application/json")
  res.json(batches);
});

module.exports = router;


