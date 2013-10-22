#include "runtime.h"

void ·Get(int32 ret) {
	ret = g->goid;
	USED(&ret);
}