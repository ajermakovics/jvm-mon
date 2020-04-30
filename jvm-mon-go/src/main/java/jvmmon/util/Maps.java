package jvmmon.util;

import java.util.LinkedHashMap;
import java.util.Map;

public class Maps {

    public static <K, V> Map<K, V> mapOf(K key, V value, Object... kvPairs) {
        Map<K, V> map = new LinkedHashMap();
        map.put(key, value);
        for (int i = 0; i < kvPairs.length - 1; i += 2) {
            map.put((K) kvPairs[i], (V) (kvPairs[i + 1]));
        }
        return map;
    }
}
