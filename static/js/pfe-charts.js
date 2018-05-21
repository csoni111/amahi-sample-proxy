let t = 6;
$(document).ready(function () {
    $(document).on('change','#timespan',function(){
        t = $(this).find("option:selected").attr('value');
        d3.json("/api/stats/?t=" + t).then(visualizePFE);
    });
    d3.json("/api/stats/?t=" + t).then(visualizePFE);
});

function visualizePFE(systemStat) {
    const stats = systemStat.stats;
    const parseTime = d3.timeParse("%s");
    stats.forEach(function(d) {
        d["timestamp"] = parseTime(d["timestamp"]);
    });
    if (stats.length > 0) {
        showGaugeChart("Ram Used", "#mem-gauge-chart", ((systemStat.total_memory - stats[0].ram_free) / systemStat.total_memory )* 100);
        showGaugeChart("Disk Used", "#disk-gauge-chart", ((systemStat.total_disk - stats[0].disk_free) / systemStat.total_disk) * 100);
        showGaugeChart("CPU Used", "#cpu-gauge-chart", stats[0].cpu_usage);
    }
}

function showGaugeChart(name, placeholder, per) {
    return c3.generate({
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
