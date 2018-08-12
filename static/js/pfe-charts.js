let t = 1;
$(document).ready(function () {
    $(document).on('change', '#timespan', function () {
        t = $(this).find("option:selected").attr('value');
        d3.json("/api/stats/?t=" + t).then(visualizePFE);
    });
    d3.json("/api/stats/?t=" + t).then(visualizePFE);
});

function visualizePFE(systemStat) {
    const stats = systemStat.stats;
    if (stats.length > 0) {
        renderGaugeCharts(stats);
        renderLineCharts(stats);
    }
}

function renderGaugeCharts(stats) {
    showGaugeChart("Ram Used", "#mem-gauge-chart", stats[0].ram_used);
    showGaugeChart("Disk Used", "#disk-gauge-chart", stats[0].disk_used);
    showGaugeChart("CPU Used", "#cpu-gauge-chart", stats[0].cpu_used);
    let memAllocated = formatBytes(stats[0].mem_alloc, 2);
    $("span.number-display", "div#mem-alloc-number").text(memAllocated[0] + " " + memAllocated[1]);
}

function showGaugeChart(name, placeholder, per) {
    c3.generate({
        bindto: placeholder,
        data: {
            columns: [
                [name, per]
            ],
            type: 'gauge',
        },
        color: {
            pattern: ['#60B044', '#F6C600', '#F97600', '#FF0000'], // the three color levels for the percentage values.
            threshold: {
                values: [30, 60, 90, 100]
            }
        },
        size: {
            height: 180
        }
    });
}

function renderLineCharts(stats) {
    const timestamp = stats.map(s => s.timestamp * 1000);
    const ramUsage = stats.map(s => s.ram_used);
    const diskUsage = stats.map(s => s.disk_used);
    const cpuUsage = stats.map(s => s.cpu_used);
    const memAllocation = stats.map(s => (s.mem_alloc / 1024 / 1024));
    showLineChart("Ram Usage", "#ram-chart", timestamp, ramUsage, "%");
    showLineChart("Disk Usage", "#disk-chart", timestamp, diskUsage, "%");
    showLineChart("CPU Usage", "#cpu-chart", timestamp, cpuUsage, "%");
    showLineChart("Memory Allocation", "#mem-alloc-chart", timestamp, memAllocation, " MB");
}

function showLineChart(name, placeholder, dataX, dataY, yLabel) {
    dataX.unshift("timestamp");
    dataY.unshift(name);
    c3.generate({
        bindto: placeholder,
        padding: {
            top: 0,
            right: 40,
            bottom: 15,
            left: 50,
        },
        data: {
            x: 'timestamp',
            xFormat: '%Q',
            columns: [dataX, dataY]
        },
        axis: {
            x: {
                type: 'timeseries',
                tick: {
                    format: '%H:%M'
                }
            },
            y : {
                tick: {
                    format: function (d) { return d + yLabel; }
                }
            }
        },
        size: {
            height: 220
        }
    });
}

function formatBytes(a, b) {
    if (0 === a) return "0 Bytes";
    let c = 1024, d = b || 2, e = ["Bytes", "KB", "MB", "GB", "TB"],
        f = Math.floor(Math.log(a) / Math.log(c));
    return [parseFloat((a / Math.pow(c, f)).toFixed(d)), e[f]]
}
