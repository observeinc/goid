// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·getg(SB), NOSPLIT, $0-4
    MOVD    g, R8
    MOVD    R8, ret+0(FP)
    RET
