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
            sb.append(quote(String.valueOf(kvPairs[i]))).append(':');
            sb.append(toJsonValue(kvPairs[i + 1]));
            if (i < kvPairs.length - 2) sb.append(",");
        }
        sb.append("}");
        return sb.toString();
    }

    public static String toJsonValue(Object value) {
        if (value == null)
            return "null";

        if (value instanceof String || value instanceof Enum || value instanceof Character)
            return quote(value.toString());

        if (value instanceof Jsonable)
            return ((Jsonable) value).toJson();

        if (value instanceof Map)
            return toJsonObject((Map) value);

        if (value instanceof Collection) { // array
            Collection<?> items = (Collection<?>) value;
            Stream<String> stream = items.stream().map(Json::toJsonValue);
            return stream.collect(joining(",", "[", "]"));
        }

        if (value instanceof Double || value instanceof Float) {
            double d = ((Number) value).doubleValue();
            if (Double.isNaN(d) || Double.isInfinite(d))
                return "0";
        }

        return value.toString();
    }

    public static <T> String toJsonObject(Map<?, T> data) {
        StringBuilder sb = new StringBuilder("{");

        Iterator<? extends Map.Entry<?, T>> iter = data.entrySet().iterator();
        while (iter.hasNext()) {
            Map.Entry<?, T> entry = iter.next();
            sb.append(quote(String.valueOf(entry.getKey()))).append(':');
            sb.append(toJsonValue(entry.getValue()));
            if (iter.hasNext()) sb.append(",");
        }

        sb.append("}");
        return sb.toString();
    }

    /** Quotes and escapes a string per RFC 8259 */
    public static String quote(String s) {
        StringBuilder sb = new StringBuilder(s.length() + 2);
        sb.append('"');
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            switch (c) {
                case '"': sb.append("\\\""); break;
                case '\\': sb.append("\\\\"); break;
                case '\n': sb.append("\\n"); break;
                case '\r': sb.append("\\r"); break;
                case '\t': sb.append("\\t"); break;
                case '\b': sb.append("\\b"); break;
                case '\f': sb.append("\\f"); break;
                default:
                    if (c < 0x20 || c == '\u2028' || c == '\u2029')
                        sb.append(String.format("\\u%04x", (int) c));
                    else
                        sb.append(c);
            }
        }
        sb.append('"');
        return sb.toString();
    }

    public static void main(String[] args) throws Exception {
        out.println(toJson("a", 1, "b", true, "c", "d\"\\\n"));
    }
}
