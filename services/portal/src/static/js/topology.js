// Derived from https://beta.observablehq.com/@mbostock/d3-force-directed-graph
let visualizationSvgId = 'voltronTopology';
let visualizationContainerId = 'topologyContainer';
let visualizationWidth = 0;
let visualizationHeight = 0;

document.addEventListener('DOMContentLoaded', function () {
    let destContainer = document.getElementById(visualizationContainerId);
    visualizationWidth = destContainer.offsetWidth;
    visualizationHeight = destContainer.offsetHeight;
    d3.json('http://voltron-sjc.cisco.com:30880/api/v1/topology')
    .then(visualizeTopology, function (e) {
        console.error('Could not obtain data!');
    });
});

function nodeColor (d) {
    return d3.interpolateCool(1 / d.group);
}

function linkColor (d) {
    return d3.interpolateSinebow(1 / d.value);
}

function drag (simulation) {
    function dragstarted(d) {
        if (!d3.event.active) simulation.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
    }

    function dragged(d) {
        d.fx = d3.event.x;
        d.fy = d3.event.y;
    }

    function dragended(d) {
        if (!d3.event.active) simulation.alphaTarget(0);
        d.fx = null;
        d.fy = null;
    }

    return d3.drag()
        .on('start', dragstarted)
        .on('drag', dragged)
        .on('end', dragended);
}

function visualizeTopology (topologyData) {
    console.debug('Topology data loaded.');
    let topologySvg = d3.select('#' + visualizationSvgId);
    let nodes = topologyData.nodes.map(d => Object.create(d));
    let links = topologyData.links.map(d => Object.create(d));
    let topologySimulation = d3.forceSimulation(nodes)
        .force('link', d3.forceLink(links).id(d => d.id))
        .force('charge', d3.forceManyBody().strength(-50))
        .force('center', d3.forceCenter(visualizationWidth / 2, visualizationHeight / 2));
    console.debug('Topology simulation started.');
    let link = topologySvg.append('g').attr('stroke', '#999').attr('stroke-opacity', 0.6)
        .selectAll('line').data(links).enter()
        .append('line').attr('stroke', linkColor);
    console.debug('Links created.');
    let node = topologySvg.append('g').attr('stroke', '#fff').attr('stroke-width', 1.5)
        .selectAll('circle').data(nodes).enter()
        .append('circle').attr('r', 7).attr('fill', nodeColor)
        .call(drag(topologySimulation));
    console.debug('Nodes created.');
    node.append('title').text(d => d.id);
    topologySimulation.on('tick', function () {
        link
            .attr('x1', d => d.source.x)
            .attr('y1', d => d.source.y)
            .attr('x2', d => d.target.x)
            .attr('y2', d => d.target.y);
        node
            .attr('cx', d => d.x)
            .attr('cy', d => d.y);
    });
    console.debug('Topology visualization setup complete.')
    return topologySvg.node();
}
