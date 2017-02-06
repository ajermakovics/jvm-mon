
var blessed = require('blessed')
 , contrib = require('blessed-contrib')
 , screen = blessed.screen()
 , grid = new contrib.grid({rows: 2, cols: 4, screen: screen})

var times = [], hist = {};

var labelStyle = {
    fg: "white",
    bold: true
}

// grid.set(row, col, rowSpan, colSpan, obj, opts)
var table = grid.set(0, 0, 1, 2, contrib.table,
 { keys: true
 , fg: 'white'
 , selectedFg: 'white'
 , selectedBg: 'blue'
 , interactive: true
 , label: 'JVM Processes'
// , width: '30%'
// , height: '30%'
 , border: {type: "line", fg: "cyan"}
 , columnSpacing: 4 //in chars
 , columnWidth: [6, 16, 10, 10, 10, 10] /*in chars*/ })


var line = grid.set(0, 2, 1, 2, contrib.line,
{
         label: "CPU %",
         maxY: 100,
         showNthLabel: 5,
         showLegend: true,
         wholeNumbersOnly: true,
         style: {
             baseline: "white",
             label: labelStyle,
             line: "yellow",
             text: "white"
         }
})

var memLine = grid.set(1, 2, 1, 2, contrib.line,
{
         label: "Heap (MB)",
         showNthLabel: 5,
         showLegend: true,
         wholeNumbersOnly: true,
         abbreviate: false,
         style: {
             baseline: "white",
             label: labelStyle,
             line: "yellow",
             text: "white"
         }
})

var cpuDonut = grid.set(1, 0, 1, 1, contrib.donut, {
    label: 'CPU',
    radius: 20,
    arcWidth: 4,
    remainColor: 'black',
    yPadding: 2,
    data: [ ]
  });

var bar = grid.set(1, 1, 1, 1, contrib.stackedBar,
       { label: 'Heap (MB)'
       , barWidth: 6
       , barSpacing: 6
       , xOffset: 5
       //, maxValue: 15
//       , height: "40%"
//       , width: "50%"
       , barBgColor: [ 'red', 'cyan']})

//----------------------
screen.append(table)
screen.append(cpuDonut)
screen.append(bar)
screen.append(line)
screen.append(memLine)

//allow control the table with the keyboard
table.focus()

screen.key(['escape', 'q', 'C-c'], function(ch, key) {
 return process.exit(0);
});

screen.key(['space'], function(ch, key) {
    addData(getData());
});

String.prototype.trunc = function(n){
    return this.substr(0, n-1);
};

var startTime = Math.floor(Date.now() / 1000);

table.rows.on("select", function (item) {
    table.selectedItem = table.rows.getItemIndex(item);
    var vm = table.vms[table.selectedItem];
    renderVmCharts(vm.Id);
    renderVmStats(vm);
    screen.render();
});

// set initial data
screen.render()

function addData(vms) {

    var tableData = {headers:["PID", "Main", "CPU", "HeapUsed", "HeapSize", "GC"], data:[]};

    for(var i = 0; i < vms.length; i++) {
        var vm = vms[i];
        var row = [vm.Id, vm.DisplayName.trunc(16), pct(vm.CpuLoad), fmt(vm.HeapUsed), fmt(vm.HeapSize), pct(vm.GcLoad)];
        tableData.data.push(row)
    }
    table.setData(tableData);
    table.vms = vms;

    var nowInSec = Math.floor(Date.now() / 1000)
    times.unshift(formatTime(nowInSec - startTime))

    for(var i = 0; i < vms.length; i++) {
        var vm = vms[i]
        var vmHist = hist[vm.Id] || {cpu: [], gc: [], size: [], used: []}
        var cpuLoad = (vm.CpuLoad * 100).toFixed(1)
        vmHist.cpu.push(cpuLoad)
        vmHist.size.push(vm.HeapSize / 1024 / 1024)
        vmHist.used.push(vm.HeapUsed / 1024 / 1024)
        vmHist.gc.push((vm.GcLoad*100).toFixed(1))
        hist[vm.Id] = vmHist
    }

    if(vms.length) {
        var selectedVm = vms[table.selectedItem || 0]
        renderVmCharts(selectedVm.Id)
        renderVmStats(selectedVm);
    }

    screen.render()
}

function formatTime(time) {
    var sec = time%60;
    return Math.floor(time/60) + ':' + ((sec<10)?'0':'') + sec;
}

function renderVmCharts(vmId) {
    var vmHist = hist[vmId]
    var cpuData = {title: 'CPU', x: times, y: vmHist.cpu, style: {line: 'yellow'}};
    var gcData = {title: 'GC', x: times, y: vmHist.gc, style: {line: 'blue'}};
    line.setData([gcData, cpuData]);

    var heapSizeData = {title: 'Size', x: times, y: vmHist.size, style: {line: 'red'}};
    var heapUsageData = {title: 'Usage', x: times, y: vmHist.used, style: {line: 'cyan'}};
    memLine.setData([heapSizeData, heapUsageData]);
}

function renderVmStats(vm) {
    cpuDonut.setData([{percent: vm.CpuLoad, label: 'CPU', 'color': 'red'}]);

    var freeHeap = vm.HeapSize - vm.HeapUsed;

    bar.setData({
       barCategory: ['Heap']
       , stackedCategory: ['Used', 'Free']
       , data:
          [ [Math.floor(vm.HeapUsed/1024/1024), Math.floor(freeHeap/1024/1024)] ]
       })
}

function fmt(bytes, decimals) {
   if(bytes == 0) return '0 B';
   var k = 1000,
       dm = decimals + 1 || 1,
       sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'],
       i = Math.floor(Math.log(bytes) / Math.log(k));
   return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

function pct(num) {
    return (num * 100).toFixed(2) + ' %'
}

setInterval(function() {
    addData(getData());
}, refreshDelay)

// "Id"
// "DisplayName"
// "HeapUsed"
// "HeapSize"
// "HeapMax"
// "NonHeapUsed"
// "NonHeapMax"
// "CpuLoad"
// "GcLoad"
// "VMVersion"
// "OSUser"
// "ThreadCount"
// "hasDeadlockThreads"
