
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
 , label: 'JVM Processes (Enter=Monitor, Del=Terminate)'
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

var threadTable = grid.set(1, 0, 1, 2, contrib.table,
 { keys: false
 , fg: 'white'
 , selectedFg: 'white'
 , selectedBg: 'blue'
 , interactive: false
 , label: 'Threads'
 , border: {type: "line", fg: "cyan"}
 , columnSpacing: 4 //in chars
 , columnWidth: [6, 16, 12, 10, 10] /*in chars*/ })
// ["TID", "Name", "State", "CPU", "Total CPU"], data:[]};

var prompt = blessed.question({
                           parent: screen,
                           border: 'line',
                           height: 'shrink',
                           width: 'half',
                           top: 'center',
                           left: 'center',
                           label: ' {blue-fg}Question{/blue-fg} (Enter=Ok, Esc=Cancel)',
                           tags: true,
                           keys: true,
                           vi: true
                         });

//----------------------
screen.append(table)
screen.append(threadTable)
screen.append(line)
screen.append(memLine)
screen.append(prompt);

//allow control the table with the keyboard
table.focus()

screen.key(['q', 'C-c'], function(ch, key) {
 return process.exit(0);
});

screen.key(['delete', 'backspace'], function(ch, key) {
    var vm = table.vms[table.rows.selected];
    var pid = vm.Id
    prompt.ask('Kill ' + vm.DisplayName.trunc(16) + ' (' + pid + ')?', function (err, val) {
        if(val)
            process.kill(pid, 'SIGTERM')
    })
    prompt.focus()
});

String.prototype.trunc = function(n){
    return this.substr(0, n-1);
};

var startTime = Math.floor(Date.now() / 1000);

table.rows.on("select", function (item) {
    table.selectedItem = table.rows.getItemIndex(item);
    var vm = table.vms[table.selectedItem];
    table.selectedVmId = vm.Id;
    renderVmCharts(vm);
    renderThreads();
    screen.render();
});

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
        renderVmCharts(selectedVm)
        renderThreads(selectedVm.threads)
        table.selectedVmId = selectedVm.Id
        threadTable.setLabel('Threads (' + selectedVm.ThreadCount + ') ')
    }

    screen.render()
}

function renderThreads(threads) {
    var tableData = {headers:["TID", "Name", "State", "CPU", "TotalCPU"], data:[]};

    if(threads) {
        for(var i = 0; i < threads.length; i++) {
            var th = threads[i];
            var row = [th.TID, th.name.trunc(16), th.state.trunc(11), pct(th.cpu), pct(th.totalCpu)];
            tableData.data.push(row)
        }
    }

    threadTable.setData(tableData);
}

function renderVmCharts(vm) {
    var vmHist = hist[vm.Id]
    var cpuData = {title: 'CPU', x: times, y: vmHist.cpu, style: {line: 'yellow'}};
    var gcData = {title: 'GC', x: times, y: vmHist.gc, style: {line: 'blue'}};
    line.setData([gcData, cpuData]);

    var heapSizeData = {title: 'Size', x: times, y: vmHist.size, style: {line: 'red'}};
    var heapUsageData = {title: 'Used ' + fmt(vm.HeapUsed), x: times, y: vmHist.used, style: {line: 'cyan'}};
    memLine.setData([heapSizeData, heapUsageData]);
    memLine.setLabel('Heap (MB), max=' + fmt(vm.HeapMax) + ', nonHeap=' + fmt(vm.NonHeapUsed));
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

function formatTime(time) {
    var sec = time%60;
    return Math.floor(time/60) + ':' + ((sec<10)?'0':'') + sec;
}

setInterval(function() {
    addData(getData(table.selectedVmId || 0));
}, refreshDelay)
