#include <array>
#include <map>
#include <sstream>
#include <string>
#include <unordered_map>
#include <vector>

#include <assert.h>
#include <ctype.h>
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <libnvpair.h>

#define stringify(s) xstringify(s)
#define xstringify(s) #s

#define fnvlist_add_double(l, n, v) assert(nvlist_add_double(l, n, v) == 0)
#define fnvlist_add_hrtime(l, n, v) assert(nvlist_add_hrtime(l, n, v) == 0)
#define arrlen(x) (sizeof(x)/sizeof(x[0]))

struct test {
	std::string name;
	std::string type;
	std::string definition;
	std::string xdr;
	std::string native;
};
std::vector<test> tests;

std::unordered_map<int, const char *> types = {
	{DATA_TYPE_BOOLEAN, "Boolean"},
	{DATA_TYPE_BYTE, "byte"},
	{DATA_TYPE_INT16, "int16"},
	{DATA_TYPE_UINT16, "uint16"},
	{DATA_TYPE_INT32, "int32"},
	{DATA_TYPE_UINT32, "uint32"},
	{DATA_TYPE_INT64, "int64"},
	{DATA_TYPE_UINT64, "uint64"},
	{DATA_TYPE_STRING, "string"},
	{DATA_TYPE_BYTE_ARRAY, "[]byte"},
	{DATA_TYPE_INT16_ARRAY, "[]int16"},
	{DATA_TYPE_UINT16_ARRAY, "[]uint16"},
	{DATA_TYPE_INT32_ARRAY, "[]int32"},
	{DATA_TYPE_UINT32_ARRAY, "[]uint32"},
	{DATA_TYPE_INT64_ARRAY, "[]int64"},
	{DATA_TYPE_UINT64_ARRAY, "[]uint64"},
	{DATA_TYPE_STRING_ARRAY, "[]string"},
	{DATA_TYPE_HRTIME, "time.Duration"},
	{DATA_TYPE_NVLIST, "interface{}"},
	{DATA_TYPE_NVLIST_ARRAY, "[]map[string]interface{}"},
	{DATA_TYPE_BOOLEAN_VALUE, "bool"},
	{DATA_TYPE_INT8, "int8"},
	{DATA_TYPE_UINT8, "uint8"},
	{DATA_TYPE_BOOLEAN_ARRAY, "[]bool"},
	{DATA_TYPE_INT8_ARRAY, "[]int8"},
	{DATA_TYPE_UINT8_ARRAY, "[]uint8"},
	{DATA_TYPE_DOUBLE, "float64"},
};

static void hexdump(const std::string str) {
	using namespace std;

	const char *bytes = str.c_str();
	const size_t length = str.size();

	if (length == 0) {
		return;
	}

	// Print bytes
	size_t pos = 0;
	char asciibytes[17];
	memset(asciibytes, '\0', 17);
	size_t end = length - (length % 16);
	for (; pos < end ; pos++) {
		unsigned i = pos % 16;
		if (i == 0) {
			// Print position
			printf("%08x ", static_cast<uint32_t>(pos));

		}
		unsigned curbyte = bytes[pos] & 0xFF;
		asciibytes[i] = isprint(curbyte) ? curbyte : '.';

		// Print current byte in hex
		printf("%02x ", curbyte);
		if (i == 7) {
			// Print current byte in hex with extra space at the end
			printf(" ");
		} else if (i == 15) {
			// Print ascii chars + endl
			printf("|%s|\n", asciibytes);
		}
	}

	memset(asciibytes, ' ', 16);
	size_t fill = 16 - (length - pos);
	for (; pos < length ; pos++) {
		unsigned i = pos % 16;
		if (i == 0) {
			// Print position
			printf("%08x ", static_cast<uint32_t>(pos));

		}
		unsigned curbyte = bytes[pos] & 0xFF;
		asciibytes[i] = isprint(curbyte) ? curbyte : '.';

		// Print current byte in hex
		printf("%02x ", curbyte);
		if (i == 7) {
			// Print current byte in hex with extra space at the end
			printf(" ");
		}
	}

	for (unsigned int i = 0 ; i < fill; i++) {
		if (i == 7) {
			// Print current byte in hex with extra space at the end
			printf(" ");
		}
		printf("   ");
	}
	if (fill) {
		printf("|%s|\n", asciibytes);
	}
}

// sanitize tags by escaping " and ` characters, and use escaped hex notation
// for non printable characters
static char *sanitize(char *name) {
	name = strdup(name);
	size_t slen = strlen(name);
	size_t buflen = strlen(name);
	// let the magic begin
	for (unsigned int i = 0; i < slen; i++) {
		assert(name[i] != '`');
		if (name[i] != '"' && isprint(name[i])) {
			continue;
		}
		if (slen + 2 >= buflen) {
			// grow the buffer
			buflen *= 2;
			char *newbuf = (char *)realloc(name, buflen);
			memset(newbuf+slen, 0, buflen-slen);
			name = newbuf;
		}
		if (name[i] == '"') {
			memmove(name+i+1, name+i, slen - i);
			name[i] = '\\';
			i+=1;
			slen+=1;
		} else {
			char ch = name[i];
			char buf[5];
			memmove(name+i+4, name+i+1, slen - i);
			sprintf(buf, "\\x%02x", ch & 0xff);
			memcpy(name+i, buf, 4);
			i+=3;
			slen+=3;
		}
	}
	return name;
}

static std::string gen_type(const char *cname) {
	std::string name("type_");
	name +=(cname);
	for (auto &&c: name) {
		if (c == ' ' || c == '(' || c == ')') {
			c = '_';
		}
	}
	return name;
}

static std::string define(nvlist_t *list, std::string type) {
	nvpair_t *pair = NULL;
	std::stringstream def;

	def << "type " << type << " struct {\n";
	char field = 'A';
	while ((pair = nvlist_next_nvpair(list, pair)) != NULL) {
		auto name = sanitize(nvpair_name(pair));
		auto type = nvpair_type(pair);
		def << "\t" << field++ << " " << types[type] << " `nv:\"" << name;
		if (type == DATA_TYPE_BYTE || type == DATA_TYPE_BYTE_ARRAY) {
			def << ",byte";
		}
		def << "\"`\n";
		free(name);
	}
	def << "}\n\n";
	return def.str();
}

static void print(test &t) {
	char *buf = NULL;
	size_t len;

	printf("\t{"
	       "name: \"%s\", "
	       "ptr: func() interface{} { return &%s{} }, "
	       "xdr: []byte(\"",
	       t.name.c_str(), t.type.c_str());

	buf = (char *)t.xdr.c_str();
	len = t.xdr.size();
	for (unsigned i = 0; i < len; i++) {
		printf("\\x%02x", buf[i] & 0xFF);
	}
	printf("\"), ");

	buf = (char *)t.native.c_str();
	len = t.native.size();
	printf("native: []byte(\"");
	for (unsigned i = 0; i < len; i++) {
		printf("\\x%02x", buf[i] & 0xFF);
	}
	printf("\")");

	printf("},\n");
}

static void pack(nvlist_t *list, const char *name) {
	char *xdr = NULL;
	char *native = NULL;
	size_t xlen;
	size_t nlen;
	int err;
	if ((err = nvlist_pack(list, &xdr, &xlen, NV_ENCODE_XDR, 0)) != 0) {
		fprintf(stderr, "nvlist_pack XDR error: %d", err);
		assert(err);
	}

	if ((err = nvlist_pack(list, &native, &nlen, NV_ENCODE_NATIVE, 0)) != 0) {
		fprintf(stderr, "nvlist_pack NATIVE error: %d", err);
		assert(err);
	}

	std::string type = gen_type(name);
	std::string def = define(list, type);

	test t;
	t.name = name;
	t.type = type;
	t.definition = def;
	t.xdr= std::string(xdr, xlen);
	t.native= std::string(native, nlen);
	tests.push_back(t);
}

static char *stra(char *s, int n) {
	size_t sl = strlen(s) + 2;
	char sep[sl];
	memset(sep, '\0', sl);
	strcat(sep, s);
	strcat(sep, ";");
	sl -= 1; // remove accounting for '\0'

	size_t size = sl * n + 1; // is really 1 byte extra but makes later stuff easier
	char *ss = (char *)calloc(1, size);
	assert(s);

	for (int i = 0; i < n; i++) {
		strcat(ss, sep);
	}
	ss[size - 2] = '\0'; //overwrite trailing ',' with '\0'
	return ss;
}

static char *stru(unsigned long long i) {
	char *s = NULL;
	int err = asprintf(&s, "%llu", i);
	if (err == -1) {
		fprintf(stderr, "asprintf error: %d\n", err);
		assert(err != -1);
	}
	return s;
}

static char *stri(long long i) {
	char *s = NULL;
	int err = asprintf(&s, "%lld", i);
	if (err == -1) {
		fprintf(stderr, "asprintf error: %d\n", err);
		assert(err != -1);
	}
	return s;
}

static char *strf(double d) {
	char *s = NULL;
	int err = asprintf(&s, "%16.17g", d);
	if (err == -1) {
		fprintf(stderr, "asprintf error: %d\n", err);
		assert(err != -1);
	}
	return s;
}

#define do_signed(lower, UPPER) do { \
	std::map<std::string, lower##_t> map; \
	map[stri(UPPER##_MIN)] = UPPER##_MIN; \
	map[stri(UPPER##_MIN+1)] = UPPER##_MIN+1; \
	map[stri(UPPER##_MIN/2)] = UPPER##_MIN/2; \
	map["-1"] = -1; \
	map["0"] = 0; \
	map["1"] = 1; \
	map[stri(UPPER##_MAX/2)] = UPPER##_MAX/2; \
	map[stri(UPPER##_MAX-1)] = UPPER##_MAX-1; \
	map[stri(UPPER##_MAX)] = UPPER##_MAX; \
	l = fnvlist_alloc(); \
	for (auto &kv : map) { \
		fnvlist_add_##lower(l, kv.first.c_str(), kv.second); \
	} \
	pack(l, stringify(lower) "s"); \
	fnvlist_free(l); \
} while(0)

#define do_unsigned(lower, UPPER) do { \
	std::map<std::string, lower##_t> map; \
	map["0"] = 0; \
	map["1"] = 1; \
	map[stru(UPPER##_MAX/2)] = UPPER##_MAX/2; \
	map[stru(UPPER##_MAX-1)] = UPPER##_MAX-1; \
	map[stru(UPPER##_MAX)] = UPPER##_MAX; \
	l = fnvlist_alloc(); \
	for (auto &kv : map) { \
		fnvlist_add_##lower(l, kv.first.c_str(), kv.second); \
	} \
	pack(l, stringify(lower) "s"); \
	fnvlist_free(l); \
} while(0)

#define do_double(lower, UPPER) do { \
	std::map<std::string, lower> map; \
	double offset = .9999;\
	map[strf(UPPER##_MIN)] = UPPER##_MIN; \
	map[strf(UPPER##_MIN/offset)] = UPPER##_MIN/offset; \
	map[strf(UPPER##_MIN/2)] = UPPER##_MIN/2; \
	map["-1"] = -1; \
	map["0"] = 0; \
	map["1"] = 1; \
	map[strf(UPPER##_MAX/2)] = UPPER##_MAX/2; \
	map[strf(UPPER##_MAX*offset)] = UPPER##_MAX*offset; \
	map[strf(UPPER##_MAX)] = UPPER##_MAX; \
	map[strf(M_E)] = M_E; \
	map[strf(M_LOG2E)] = M_LOG2E; \
	map[strf(M_LOG10E)] = M_LOG10E; \
	map[strf(M_LN2)] = M_LN2; \
	map[strf(M_LN10)] = M_LN10; \
	map[strf(M_PI)] = M_PI; \
	map[strf(M_PI_2)] = M_PI_2; \
	map[strf(M_PI_4)] = M_PI_4; \
	map[strf(M_1_PI)] = M_1_PI; \
	map[strf(M_2_SQRTPI)] = M_2_SQRTPI; \
	map[strf(M_SQRT2)] = M_SQRT2; \
	map[strf(M_SQRT1_2)] = M_SQRT1_2; \
	l = fnvlist_alloc(); \
	for (auto &kv : map) { \
		fnvlist_add_##lower(l, kv.first.c_str(), kv.second); \
	} \
	pack(l, stringify(lower) "s"); \
	fnvlist_free(l); \
} while(0)

#define arrset(array, alen, val) do { \
	size_t i = 0; \
	for (i = 0; i < alen; i++) { \
		array[i] = val; \
	} \
} while (0)

#define do_signed_array(len, lower, UPPER) do { \
	typedef std::array<lower##_t, len> array; \
	array arr; \
	std::map<std::string, array> map; \
	arr.fill(UPPER##_MIN); map[stra(stri(UPPER##_MIN), arr.size())] = arr; \
	arr.fill(UPPER##_MIN+1); map[stra(stri(UPPER##_MIN+1), arr.size())] = arr; \
	arr.fill(UPPER##_MIN/2); map[stra(stri(UPPER##_MIN/2), arr.size())] = arr; \
	arr.fill(-1), map[stra(stri(-1), len)] = arr; \
	arr.fill(0); map[stra(stri(0), arr.size())] = arr; \
	arr.fill(1); map[stra(stri(1), arr.size())] = arr; \
	arr.fill(UPPER##_MAX/2); map[stra(stri(UPPER##_MAX/2), arr.size())] = arr; \
	arr.fill(UPPER##_MAX-1); map[stra(stri(UPPER##_MAX-1), arr.size())] = arr; \
	arr.fill(UPPER##_MAX); map[stra(stri(UPPER##_MAX), arr.size())] = arr; \
	l = fnvlist_alloc(); \
	for (auto &kv : map) { \
		fnvlist_add_##lower##_array(l, kv.first.c_str(), kv.second.data(), kv.second.size()); \
	} \
	pack(l, stringify(lower) " array(" stringify(len)")"); \
	fnvlist_free(l); \
} while(0) \

#define do_unsigned_array(len, lower, UPPER) do { \
	typedef std::array<lower##_t, len> array; \
	array arr; \
	std::map<std::string, array> map; \
	arr.fill(0); map[stra(stru(0), arr.size())] = arr; \
	arr.fill(1); map[stra(stru(1), arr.size())] = arr; \
	arr.fill(UPPER##_MAX/2); map[stra(stru(UPPER##_MAX/2), arr.size())] = arr; \
	arr.fill(UPPER##_MAX-1); map[stra(stru(UPPER##_MAX-1), arr.size())] = arr; \
	arr.fill(UPPER##_MAX); map[stra(stru(UPPER##_MAX), arr.size())] = arr; \
	l = fnvlist_alloc(); \
	for (auto &kv : map) { \
		fnvlist_add_##lower##_array(l, kv.first.c_str(), kv.second.data(), kv.second.size()); \
	} \
	pack(l, stringify(lower) " array(" stringify(len)")"); \
	fnvlist_free(l); \
} while(0) \

#define do_string_array(list, array) do { \
	size_t nelems = arrlen(array); \
	std::string name; \
	for (size_t i = 0; i < nelems - 1; i++) { \
		name += array[i]; \
		name += ";"; \
	} \
	name += array[nelems - 1]; \
	fnvlist_add_string_array(list, name.c_str(), (char **)array, nelems); \
	pack(list, "string array"); \
} while(0)

int main(int argc, char** argv) {
	nvlist_t *l = NULL;
	{
		l = fnvlist_alloc();
		pack(l, "empty");
		fnvlist_free(l);
	}

	{
		l = fnvlist_alloc();
		fnvlist_add_boolean(l,"true");
		pack(l,"boolean");
		fnvlist_free(l);
	}

	{
		l = fnvlist_alloc();
		fnvlist_add_boolean_value(l, "false", B_FALSE);
		fnvlist_add_boolean_value(l, "true", B_TRUE);
		pack(l, "bools");
		fnvlist_free(l);
	}
	{
		l = fnvlist_alloc();
		size_t nelems = 5;
		boolean_t array[nelems];

		arrset(array, nelems, B_FALSE);
		fnvlist_add_boolean_array(l, stra("false", nelems), array, nelems);

		arrset(array, nelems, B_TRUE);
		fnvlist_add_boolean_array(l, stra("true", nelems), array, nelems);

		pack(l, "bool array");
		fnvlist_free(l);
	}

	{
		l = fnvlist_alloc();
		//fnvlist_add_byte(l, "-128", -128);
		fnvlist_add_byte(l, "0", 0);
		fnvlist_add_byte(l, "1", 1);
		fnvlist_add_byte(l, "127", 127);
		pack(l, "bytes");
		fnvlist_free(l);
	}

	{
		l = fnvlist_alloc();
		size_t nelems = 5;
		unsigned char array[nelems];

		//arrset(array, nelems, -128);
		//fnvlist_add_byte_array(l, stra("-128", nelems), array, nelems);

		arrset(array, nelems, 0);
		fnvlist_add_byte_array(l, stra("0", nelems), array, nelems);

		arrset(array, nelems, 1);
		fnvlist_add_byte_array(l, stra("1", nelems), array, nelems);

		arrset(array, nelems, 127);
		fnvlist_add_byte_array(l, stra("127", nelems), array, nelems);

		pack(l, "byte array");
		fnvlist_free(l);
	}


	do_signed(int8, INT8);
	do_signed(int16, INT16);
	do_signed(int32, INT32);
	do_signed(int64, INT64);
	do_signed_array(4, int8, INT8);
	do_signed_array(4, int16, INT16);
	do_signed_array(4, int32, INT32);
	do_signed_array(4, int64, INT64);
	do_signed_array(5, int8, INT8);
	do_signed_array(5, int16, INT16);
	do_signed_array(5, int32, INT32);
	do_signed_array(5, int64, INT64);

	do_unsigned(uint8, UINT8);
	do_unsigned(uint16, UINT16);
	do_unsigned(uint32, UINT32);
	do_unsigned(uint64, UINT64);
	do_unsigned_array(4, uint8, UINT8);
	do_unsigned_array(4, uint16, UINT16);
	do_unsigned_array(4, uint32, UINT32);
	do_unsigned_array(4, uint64, UINT64);
	do_unsigned_array(5, uint8, UINT8);
	do_unsigned_array(5, uint16, UINT16);
	do_unsigned_array(5, uint32, UINT32);
	do_unsigned_array(5, uint64, UINT64);

	l = fnvlist_alloc();
	fnvlist_add_string(l, "0", "0");
	fnvlist_add_string(l, "01", "01");
	fnvlist_add_string(l, "012", "012");
	fnvlist_add_string(l, "0123", "0123");
	fnvlist_add_string(l, "01234", "01234");
	fnvlist_add_string(l, "012345", "012345");
	fnvlist_add_string(l, "0123456", "0123456");
	fnvlist_add_string(l, "01234567", "01234567");
	fnvlist_add_string(l, "\xff\"", "\xff\"");
	pack(l, "strings");
	fnvlist_free(l);

	{
		const char *array[] = {
			"0",
			"01",
			"012",
			"0123",
			"01234",
			"012345",
			"0123456",
			"01234567",
			"\xff\"",
		};
		l = fnvlist_alloc();
		do_string_array(l, array);
		fnvlist_free(l);
	}
	do_signed(hrtime, INT64);

	{
		l = fnvlist_alloc();
		nvlist_t *ll = fnvlist_alloc();
		fnvlist_add_boolean_value(ll, "false", B_FALSE);
		fnvlist_add_boolean_value(ll, "true", B_TRUE);
		fnvlist_add_nvlist(l, "2", ll);
		fnvlist_free(ll);

		ll = fnvlist_alloc();
		fnvlist_add_uint8(ll, "0", 0);
		fnvlist_add_uint8(ll, "1", 1);
		fnvlist_add_boolean_value(ll, "false", B_FALSE);
		fnvlist_add_boolean_value(ll, "true", B_TRUE);
		fnvlist_add_nvlist(l, "4", ll);
		pack(l, "nvlist");
		fnvlist_free(ll);

		fnvlist_free(l);
	}
	{
		std::string name;
		nvlist_t *l1 = fnvlist_alloc();
		fnvlist_add_boolean_value(l1, "false", B_FALSE);
		fnvlist_add_boolean_value(l1, "true", B_TRUE);
		name += std::to_string(fnvlist_num_pairs(l1));
		name += ";";

		nvlist_t *l2 = fnvlist_alloc();
		fnvlist_add_uint8(l2, "1", 1);
		fnvlist_add_uint8(l2, "2", 2);
		fnvlist_add_uint8(l2, "3", 3);
		name += std::to_string(fnvlist_num_pairs(l2));

		nvlist_t *larr[] = {l1, l2};
		size_t nelems = arrlen(larr);

		l = fnvlist_alloc();
		fnvlist_add_nvlist_array(l, name.c_str(), larr, nelems);
		pack(l, "nvlist array");
		fnvlist_free(l2);
		fnvlist_free(l1);
		fnvlist_free(l);
	}

	do_double(double, DBL);

	{
		l = fnvlist_alloc();
		fnvlist_add_int16_array(l, "empty int16", {}, 0);
		fnvlist_add_int32_array(l, "empty int32", {}, 0);
		fnvlist_add_int64_array(l, "empty int64", {}, 0);
		fnvlist_add_int8_array(l, "empty int8", {}, 0);
		fnvlist_add_nvlist_array(l, "empty nvlist", {}, 0);
		fnvlist_add_string_array(l, "empty string", {}, 0);
		fnvlist_add_uint16_array(l, "empty uint16", {}, 0);
		fnvlist_add_uint32_array(l, "empty uint32", {}, 0);
		fnvlist_add_uint64_array(l, "empty uint64", {}, 0);
		fnvlist_add_uint8_array(l, "empty uint8", {}, 0);
		pack(l, "empty arrays");
		fnvlist_free(l);
	}

	char *testrun = getenv("NV_TEST_RUN");
	bool xdr = false;
	bool native = false;
	for (int i = 0; i < argc; i++) {
		if (!(strcmp(argv[i], "-x") && strcmp(argv[i], "--xdr"))) {
			xdr = true;
		}
		if (!(strcmp(argv[i], "-n") && strcmp(argv[i], "--native"))) {
			native = true;
		}
	}
	if (xdr || native) {
		for (test &t: tests) {
			if (testrun && t.name != testrun) {
				continue;
			}
			printf("%s ", t.name.c_str());
			if (xdr) {
				printf("xdr:\n");
				hexdump(t.xdr);
			}
			if (native) {
				printf("native:\n");
				hexdump(t.native);
			}
			printf("\n");
		}
		return 0;
	}

	printf("package nv\n"
	       "\n"
	       "/* !!! GENERATED FILE DO NOT EDIT !!! */\n"
	       "\n"
	       "import \"time\"\n\n");

	for (test &t: tests) {
		printf("%s", t.definition.c_str());
	}

	printf("var good = []struct {\n"
	       "\tname   string\n"
	       "\tptr    func() interface{}\n"
	       "\txdr    []byte\n"
	       "\tnative []byte\n"
	       "}{\n");

	for (test &t: tests) {
		print(t);
	}
	printf("}\n");

	return 0;
}
