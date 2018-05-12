$(document).ready(function () {
    d3.json("./fs.json").then(visualize)
});

function visualize(allFs) {
    // let records = recordsJson;
    // const dateFormat = d3.time.format("%s");

    // records.forEach(function(d) {
    //     d["timestamp"] = dateFormat.parse(d["timestamp"]);
    // });


    const ndx = crossfilter(allFs);

    const all = ndx.groupAll();

    const numberOfFS = dc.numberDisplay("#unique-fs-number");

    numberOfFS
        .formatNumber(d3.format("d"))
        .valueAccessor((d) => { return d; })
        .group(all);

    dc.renderAll();
}
