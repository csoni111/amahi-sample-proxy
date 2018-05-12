$(document).ready(function () {
    d3.json("./fs.json").then(visualizeFS)
    d3.json("./connections.json").then(visualizeConn)
});

function visualizeFS(allFs) {
    // create crossfilter instance
    const ndx = crossfilter(allFs);

    // define data dimensions
    const versionDim = ndx.dimension((d) => { return d.version; });
    const archDim = ndx.dimension((d) => { return d.arch; });

    // define data groups
    const versionGroup = versionDim.group();
    const archGroup = archDim.group();
    const all = ndx.groupAll();

    // prepare charts
    const versionPie = dc.pieChart("#version-pie");
    const archPie = dc.pieChart("#arch-pie");
    const numberOfFS = dc.numberDisplay("#unique-fs-number");

    versionPie
        .slicesCap(4)
        .innerRadius(35)
        .dimension(versionDim)
        .group(versionGroup)
        .legend(dc.legend())
        .on('pretransition', function(chart) {
            chart.selectAll('text.pie-slice').text(function(d) {
                return d.data.key + ' (' + dc.utils.printSingleValue((d.endAngle - d.startAngle) / (2*Math.PI) * 100) + '%)';
            })
        });

    archPie
        .slicesCap(4)
        .innerRadius(35)
        .dimension(archDim)
        .group(archGroup)
        .legend(dc.legend())
        .on('pretransition', function(chart) {
            chart.selectAll('text.pie-slice').text(function(d) {
                return d.data.key + ' (' + dc.utils.printSingleValue((d.endAngle - d.startAngle) / (2*Math.PI) * 100) + '%)';
            })
        });

    numberOfFS
        .formatNumber(d3.format("d"))
        .valueAccessor((d) => { return d; })
        .group(all);

    dc.renderAll();
}

function visualizeConn(connections) {
    const parseTime = d3.timeParse("%s");
    connections.forEach(function(d) {
        d["timestamp"] = parseTime(d["timestamp"]);
        d["timestamp"].setSeconds(0);
    });
    console.log(connections);

    // create crossfilter instance
    const ndx = crossfilter(connections);

    // define data dimensions
    const dateDim = ndx.dimension(function(d) { return d.timestamp; });

    // define data groups
    const numConnectionsByDate = dateDim.group();

    const minDate = dateDim.bottom(1)[0]["timestamp"];
    const maxDate = dateDim.top(1)[0]["timestamp"];

    // prepare charts
    var connChart = dc.barChart("#conn-chart");

    connChart
        .margins({top: 10, right: 50, bottom: 20, left: 20})
        .dimension(dateDim)
        .group(numConnectionsByDate)
        .transitionDuration(500)
        .x(d3.scaleTime().domain([minDate, maxDate]))
        .elasticY(true)
        .yAxis().ticks(4);

    dc.renderAll();
}
