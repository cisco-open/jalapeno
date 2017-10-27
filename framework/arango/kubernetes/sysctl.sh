echo 2 > /proc/sys/vm/overcommit_memory
sysctl -w "vm.max_map_count=128000"
echo madvise > /sys/kernel/mm/transparent_hugepage/enabled
echo madvise > /sys/kernel/mm/transparent_hugepage/defrag
