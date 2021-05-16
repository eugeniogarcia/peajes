function batch(id,procesados, fallados, pendientes){
    return {
        batch: id,
        processed_records: procesados,
        failed_records: fallados,
        un_processed_records: pendientes,
        actualiza: function(error) {
            if (this.un_processed_records>0){
                this.un_processed_records--;
                error ? this.failed_records++ : this.processed_records++;
            } 
        }
    }
}

var num_batches=20;
var intervalo=10;

var batches = new Array(num_batches);
var elapsed_batches = new Array(num_batches);

for (var i=0;i<num_batches;i++){
    batches[i]=batch(1000+i,0,0,1000);
    elapsed_batches[i]=0;
}

//Probabilidad de que falle en %
probabilidad_fallo=5;
//Probabilidad de que cancele en por mil
probabilidad_cancelar = 5;
//Duracion media en segundos
probabilidad_duracion_media = 30;

setInterval(function () {
    for (var i=0;i<num_batches;i++){
        //Acumula tiempo de procesamiento
        elapsed_batches[i]+=intervalo;

        //Simula la cancelación
        if (Math.floor(Math.random() * 1000) < probabilidad_cancelar) 
        elapsed_batches[i]=-1000000;
        
        //Simula el procesamiento
        //Si se ha procesado ya...
        if (Math.floor(Math.random() * 2 * probabilidad_duracion_media) < elapsed_batches[i]){
            //Resetea
            elapsed_batches[i]=0;
            //Puede haber sido un error o un acierto
            batches[i].actualiza(Math.floor(Math.random() * 100) < probabilidad_fallo ? true : false);
        }
    }
}, intervalo * 1000);

module.exports=batches;