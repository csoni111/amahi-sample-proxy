let t = 6;
$(document).ready(function () {
    $(document).on('change','#timespan',function(){
        t = $(this).find("option:selected").attr('value');
        d3.json("/api/connections/?t=" + t).then(visualizeFS);
    });
    d3.json("/api/connections/").then(visualizeFS);
});

function visualizeFS(connections) {
    const parseTime = d3.timeParse("%s");
    connections.forEach(function(d) {
        d["timestamp"] = parseTime(d["timestamp"]);
        d["timestamp"].setSeconds(0);
    });

    // create crossfilter instance
    const ndx = crossfilter(connections);

    // define data dimensions
    const versionDim = ndx.dimension((d) => { return d.fs_info.version; });
    const archDim = ndx.dimension((d) => { return d.fs_info.arch; });
    const dateDim = ndx.dimension(function(d) { return d.timestamp; });

    // define data groups
    const versionGroup = versionDim.group();
    const archGroup = archDim.group();
    const all = ndx.groupAll();
    const numConnectionsByDate = dateDim.group();
    const maxDate = new Date();
    const minDate = new Date(maxDate.getTime() - (t * 60 * 60 * 1000)); // last t hours

    // prepare charts
    const versionPie = dc.pieChart("#version-pie");
    const archPie = dc.pieChart("#arch-pie");
    const numberOfFS = dc.numberDisplay("#unique-fs-number");
    const connChart = dc.lineChart("#conn-chart");

    connChart
        .margins({top: 10, right: 50, bottom: 20, left: 20})
        .dimension(dateDim)
        .group(numConnectionsByDate)
        .transitionDuration(500)
        .x(d3.scaleTime().domain([minDate, maxDate]))
        .elasticY(true)
        .yAxis().ticks(4);

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
