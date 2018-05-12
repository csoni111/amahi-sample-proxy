$(document).ready(function () {
    d3.json("./fs.json").then(visualize)
});

function visualize(allFs) {
    // let records = recordsJson;
    // const dateFormat = d3.time.format("%s");

    // records.forEach(function(d) {
    //     d["timestamp"] = dateFormat.parse(d["timestamp"]);
    // });

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
