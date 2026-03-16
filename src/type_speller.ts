import type { RecordKey, RecordLocation, ResolvedType } from "skir-internal";
import { getClassName, modulePathToAlias } from "./naming.js";

/**
 * Transforms a type found in a `.skir` file into a Go type.
 */
export class TypeSpeller {
  constructor(
    readonly recordMap: ReadonlyMap<RecordKey, RecordLocation>,
    readonly modulePath: string,
  ) {}

  getGoType(type: ResolvedType): string {
    switch (type.kind) {
      case "record": {
        const recordLocation = this.recordMap.get(type.key)!;
        const className = getClassName(recordLocation);
        if (recordLocation.modulePath === this.modulePath) {
          return className;
        } else {
          const packageAlias = modulePathToAlias(recordLocation.modulePath);
          return `${packageAlias}.${className}`;
        }
      }
      case "array": {
        const itemType = this.getGoType(type.item);
        return `skir_client.Array[${itemType}]`;
      }
      case "optional": {
        const otherType = this.getGoType(type.other);
        return `*${otherType}`;
      }
      case "primitive": {
        const { primitive } = type;
        switch (primitive) {
          case "bool":
          case "int32":
          case "int64":
          case "float32":
          case "float64":
          case "string":
            return primitive;
          case "hash64":
            return "uint64";
          case "timestamp":
            return "time.Time";
          case "bytes":
            return "skir_client.Bytes";
        }
      }
    }
  }

  getClassName(recordKey: RecordKey): string {
    const record = this.recordMap.get(recordKey)!;
    return getClassName(record);
  }

  getSerializerExpression(type: ResolvedType): string {
    switch (type.kind) {
      case "primitive": {
        switch (type.primitive) {
          case "bool":
            return "build.skir.Serializers.bool";
          case "int32":
            return "build.skir.Serializers.int32";
          case "int64":
            return "build.skir.Serializers.int64";
          case "hash64":
            return "build.skir.Serializers.hash64";
          case "float32":
            return "build.skir.Serializers.float32";
          case "float64":
            return "build.skir.Serializers.float64";
          case "timestamp":
            return "build.skir.Serializers.timestamp";
          case "string":
            return "build.skir.Serializers.string";
          case "bytes":
            return "build.skir.Serializers.bytes";
        }
        const _: never = type.primitive;
        throw TypeError();
      }
      case "array": {
        if (type.key) {
          const keyChain = type.key.path.map((f) => f.name.text).join(".");
          const path = type.key.path.map((f) => "TODO").join(".");
          return (
            "build.skir.internal.keyedListSerializer(\n" +
            this.getSerializerExpression(type.item) +
            `,\n"${keyChain}",\n{ it.${path} },\n)`
          );
        } else {
          return (
            "build.skir.Serializers.list(\n" +
            this.getSerializerExpression(type.item) +
            ",\n)"
          );
        }
      }
      case "optional": {
        return (
          `build.skir.Serializers.optional(\n` +
          this.getSerializerExpression(type.other) +
          `,\n)`
        );
      }
      case "record": {
        return this.getClassName(type.key) + ".serializer";
      }
    }
  }

  getGoSerializerExpression(type: ResolvedType): string {
    switch (type.kind) {
      case "primitive": {
        switch (type.primitive) {
          case "bool":
            return "skir_client.BoolSerializer()";
          case "int32":
            return "skir_client.Int32Serializer()";
          case "int64":
            return "skir_client.Int64Serializer()";
          case "hash64":
            return "skir_client.Hash64Serializer()";
          case "float32":
            return "skir_client.Float32Serializer()";
          case "float64":
            return "skir_client.Float64Serializer()";
          case "timestamp":
            return "skir_client.TimestampSerializer()";
          case "string":
            return "skir_client.StringSerializer()";
          case "bytes":
            return "skir_client.BytesSerializer()";
        }
        const _: never = type.primitive;
        throw TypeError();
      }
      case "array": {
        return (
          "skir_client.ArraySerializer(\n" +
          this.getGoSerializerExpression(type.item) +
          ",\n)"
        );
      }
      case "optional": {
        return (
          "skir_client.OptionalSerializer(\n" +
          this.getGoSerializerExpression(type.other) +
          ",\n)"
        );
      }
      case "record": {
        const recordLocation = this.recordMap.get(type.key)!;
        const className = getClassName(recordLocation);
        if (recordLocation.modulePath === this.modulePath) {
          return `${className}_serializer()`;
        } else {
          const packageAlias = modulePathToAlias(recordLocation.modulePath);
          return `${packageAlias}.${className}_serializer()`;
        }
      }
    }
  }
}
