---
title: "Inicio"
---

# Febrero 2025

## Indicaciones para Emisión de Presupuesto 2025

Las siguientes indicaciones aplican para cualquier órgano o asociación federada
plena. Favor realizar la solicitud a través del sistema, o como segunda opción enviando
un documento al correo finanzas.feucr@ucr.ac.cr, en cuyo caso se registrará
manualmente en el sistema para darle seguimiento.

- **Fecha límite de entrega de servicios**: 25 de febrero, extendible mediante solicitud de prórroga dirigida a la Contraloría Estudiantil (en cuyo caso debe enviarse antes del 26 de febrero y se permite prorrogarse la entrega de solicitudes hasta el 7 de marzo).
- **Fecha límite de entrega de activos**: 24 de febrero, no extendible.
- **Detalle**: Redactar el detalle de lo que se requiere, por ejemplo, fecha o rango de fechas, desglose del servicio, especificaciones del activo, etc.
- **Justificación**: Redactar siguiendo las [indicaciones emitidas por la Contraloría Estudiantil](#lineamientos-coes).

Tanto para las asociaciones federadas plenas como para los órganos, se
establece que la emisión de la partida de activos deberá realizarse
exclusivamente durante el primer semestre del año I-2025, en cumplimiento con
las recomendaciones recibidas por la Oficina de Suministros (OSUM) para acatar
estas indicaciones, se insta a las asociaciones estudiantiles planificar el uso
del presupuesto de activos a más tardar al finalizar el mes de febrero.

Asimismo, para las solicitudes de activos y servicios con motivo de
alimentación, serigrafiado y entretenimiento, que se puedan prever o estén
planificadas antes de la fecha límite, será necesario presentar las solicitudes
de manera anticipada a la Secretaría de Finanzas, a más tardar el 25 de febrero.

Esto permitirá fomentar una gestión más eficiente y ordenada de las solicitudes,
garantizando una adecuada planificación y asignación de los recursos
disponibles.

### Lineamientos COES

Emitido por Contraloría Estudiantil para la elaboración de justificaciones:

1. Detallar para qué es el servicio que se necesita: Si es una cafeteada, almuerzo, etc.
2. Dónde se va a llevar esta actividad y si ya cuentan con el permiso de la misma: Por  ejemplo, _"En el segundo piso de la facultad y se cuenta con el permiso del decanato"_.
3. Fecha o rango de fechas, nombre de la actividad o evento conmemorativo, cultural, etc.
4. Para quienes van dirigido el servicio. Si es abierto a toda la población estudiantil no es necesario especificar el método de selección, pero si es sólo para 40 personas, por ejemplo, indicar que se seleccionaron a esas personas por medio de un formulario, afiches o el método que corresponda.
5. ¿Porqué se necesita este servicio?.
6. ¿Por qué es necesaria la cantidad solicitada?. Por ejemplo, si es alimentación, indicar la razón por la que se pide la cantidad de unidades especificada.
7. ¿De qué forma este servicio beneficia a estas personas?
8. Los detalles del servicio: en caso de que sea de alimentos deberán especificar si es un casado, una lasaña, un  pastel, etc.  Si  lleva  refresco, fruta, postre, entre otros detalles.
9. Declaración jurada: _"La contratación de este servicio será realizado bajo la normativa nacional e institucional de la UCR velando por el correcto cumplimiento de los deberes y derechos de ambas partes involucradas"_.

# Enero 2025

## Distribución Presupuestaria 2025

<canvas style="padding:0.2em 1em;" id="dist2025cse"></canvas>
<script>
const dist2025cse = document.getElementById('dist2025cse');
new Chart(dist2025cse, {
    type: 'bar',
    data: {
        labels: [
            "Cuerpo Coordinador",
            "Asociaciones Estudiantiles Federadas Plenas de la Sede Rodrigo Facio",
            "Asociaciones Estudiantiles de Sedes Regionales",
            "Consejos de Asociaciones de Carrera de Sedes y Recintos Regionales",
            "Comisión Evaluadora de Proyectos",
            "Consejos de Asociaciones Estudiantiles",
            "Consejo de Estudiantes de Sedes y Recintos Regionales",
        ],
        datasets: [{
            label: "Presupuesto solamente CSE 2025 (₡)",
            data: [
                7244939.73,
                75347373.14,
                28979758.90,
                17387855.34,
                13040891.51,
                5795951.78, 
                5795951.78,
            ],
            borderWidth: 1
        }]
    },
    options: {
        indexAxis: 'y',
        scales: {
            y: {
                beginAtZero: true
            },
            x: {
                display: false
            }
        },
    }
});
</script>

<canvas style="padding:0.2em 1em;" id="dist2025org"></canvas>
<script>
const dist2025org = document.getElementById('dist2025org');
new Chart(dist2025org, {
    type: 'bar',
    data: {
        labels: [
            "Tribunal Electoral Estudiantil Universitario",
            "Contraloría Estudiantil",
            "Defensoría Estudiantil Universitaria",
            "Frente Ecologista Universitario",
            "Editorial Estudiantil",
            "Secretaría de Finanzas",
            "Procuraduría Estudiantil Universitaria",
        ],
        datasets: [{
            label: "Presupuesto Órganos (sin CSE ni DIR) 2025 (₡)",
            data: [
                24632795.07,
                2897975.89,
                5795951.78,
                1448987.95,
                11591903.56,
                1448987.95,
                1448987.95
            ],
            borderWidth: 1
        }]
    },
    options: {
        indexAxis: 'y',
        scales: {
            y: {
                beginAtZero: true
            },
            x: {
                display: false
            }
        },
    }
});
</script>
