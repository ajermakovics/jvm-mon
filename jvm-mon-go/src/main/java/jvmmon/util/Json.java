package jvmmon.util;

import jvmmon.model.Jsonable;

import java.util.*;
import java.util.stream.Stream;

import static java.lang.System.out;
import static java.util.stream.Collectors.joining;

/** Json generator **/
public class Json {

    public static String toJson(Object... kvPairs) {
        StringBuilder sb = new StringBuilder("{");
        for (int i = 0; i < kvPairs.length - 1; i += 2) {
            sb.append('"').append(kvPairs[i]).append("\":");
            sb.append(toJsonValue(kvPairs[i + 1]));
            if(i < kvPairs.length - 2) sb.append(",");
        }
        sb.append("}");
        return sb.toString();
    }

    public static String toJsonValue(Object value) {
        if(value == null)
            return null;

        if (value instanceof String || value instanceof Enum)
            return "\"" + value + "\"";

        if (value instanceof Jsonable)
            return ((Jsonable) value).toJson();

        if (value instanceof Map)
            return toJsonObject((Map) value);

        if (value instanceof Collection) { // array
            Collection items = (Collection) value;
            Stream<String> stream = items.stream().map(Json::toJsonValue);
            return stream.collect(joining(",", "[", "]"));
        }

        return value.toString();
    }

    public static <T> String toJsonObject(Map<String, T> data) {
        StringBuilder sb = new StringBuilder("{");

        Iterator<Map.Entry<String, T>> iter = data.entrySet().iterator();
        while (iter.hasNext()) {
            Map.Entry<String, T> entry = iter.next();
            sb.append('"').append(entry.getKey()).append('"').append(':');
            sb.append(toJsonValue(entry.getValue()));
            if (iter.hasNext()) sb.append(",");
        }

        sb.append("}");
        return sb.toString();
    }

    public static void main(String[] args) throws Exception {
        out.println(toJson("a", 1, "b", true, "c", "d"));
    }
}
