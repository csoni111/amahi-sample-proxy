const FS = {charts: {}, t: 1, ndx: null};

$(document).ready(function () {
    $(document).on('change', '#timespan', function () {
        FS.t = $(this).find("option:selected").attr('value');
        updateCharts();
        d3.json("/api/connections/?t=" + FS.t).then(visualizeFS);
    });
    loadCharts();
    setInterval(updateCharts, 5000);
});

function loadCharts() {
    d3.json("/api/fs/").then(presentFSStats);
    d3.json("/api/connections/?t=" + FS.t).then(visualizeFS);
}

function updateCharts() {
    // d3.json("/api/connections/?t=" + FS.t).then(updateCrossFilter);
    d3.json("/api/fs/").then(presentFSStats);
}

function presentFSStats(stat) {
    $("span.number-display", "div#connected-fs-number").text(stat.total_connected);
    $("span.number-display", "div#streaming-fs-number").text(stat.total_streaming);
}

function preformatData(connections) {
    const parseTime = d3.timeParse("%s");
    connections.forEach(function (d) {
        d["timestamp"] = parseTime(d["timestamp"]);
        d["timestamp"].setSeconds(0);
    });
}

function visualizeFS(connections) {
    preformatData(connections);

    // create crossfilter instance
    FS.ndx = crossfilter(connections);

    // define data dimensions
    const versionDim = FS.ndx.dimension((d) => {
        return d.fs_info.version;
    });
    const archDim = FS.ndx.dimension((d) => {
        return d.fs_info.arch;
    });
    const dateDim = FS.ndx.dimension(function (d) {
        return d.timestamp;
    });

    // define data groups
    const versionGroup = versionDim.group();
    const archGroup = archDim.group();
    const all = FS.ndx.groupAll();
    const numConnectionsByDate = dateDim.group();
    const maxDate = new Date();
    const minDate = new Date(maxDate.getTime() - (FS.t * 60 * 60 * 1000)); // last t hours

    // prepare charts
    FS.charts.versionPie = dc.pieChart("#version-pie");
    FS.charts.archPie = dc.pieChart("#arch-pie");
    FS.charts.numberOfFS = dc.numberDisplay("#unique-fs-number");
    FS.charts.connChart = dc.lineChart("#conn-chart");

    FS.charts.connChart
        .margins({top: 10, right: 50, bottom: 20, left: 20})
        .dimension(dateDim)
        .group(numConnectionsByDate)
        .transitionDuration(500)
        .x(d3.scaleTime().domain([minDate, maxDate]))
        .elasticY(true)
        .yAxis().ticks(4);

    FS.charts.versionPie
        .slicesCap(4)
        .innerRadius(35)
        .dimension(versionDim)
        .group(versionGroup)
        .legend(dc.legend())
        .on('pretransition', function (chart) {
            chart.selectAll('text.pie-slice').text(function (d) {
                return d.data.key + ' (' + getAnglePer(d) + '%)';
            })
        });

    FS.charts.archPie
        .slicesCap(4)
        .innerRadius(35)
        .dimension(archDim)
        .group(archGroup)
        .legend(dc.legend())
        .on('pretransition', function (chart) {
            chart.selectAll('text.pie-slice').text(function (d) {
                return d.data.key + ' (' + getAnglePer(d) + '%)';
            })
        });

    FS.charts.numberOfFS
        .formatNumber(d3.format("d"))
        .valueAccessor((d) => {
            return d;
        })
        .group(all);

    dc.renderAll();
}

function getAnglePer(d) {
    return dc.utils.printSingleValue((d.endAngle - d.startAngle) / (2 * Math.PI) * 100);
}

/*
function updateCrossFilter(newConnections) {
    if (FS.ndx == null) {
        return;
    }
    preformatData(newConnections);
    let filters = {};
    for (let name in FS.charts) {
        filters[name] = FS.charts[name].filter();
        FS.charts[name].filter(null);
    }
    FS.ndx.remove();
    FS.ndx.add(newConnections);
    for (let name in FS.charts) {
        if (filters[name] !== null) {
            FS.charts[name].filter([filters[name]]);
        }
    }
    dc.redrawAll();
}*/
