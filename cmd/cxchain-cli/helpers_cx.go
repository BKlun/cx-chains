package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	cxcore "github.com/skycoin/cx/cx"
	"github.com/skycoin/cx/cxgo/actions"
	"github.com/skycoin/cx/cxgo/cxgo0"
	"github.com/skycoin/cx/cxgo/cxlexer"
	"github.com/skycoin/cx/cxgo/cxprof"
)

/*
	<<< HELPERS >>>
*/

// StartPreparationProfiling starts profiling for preparation steps.
// TODO @evanlinjin: We should re-enable this at a later stage.
func StartPreparationProfiling(funcName string, debugLexer bool, debugProf int) (logrus.FieldLogger, func()) {
	log := log.WithField("func", funcName)

	// Start CPU profiling.
	stopCPUProf, err := cxprof.StartCPUProfile(funcName, debugProf)
	if err != nil {
		log.WithError(err).Error("Failed to start CPU profiling.")
	}

	// Start Log profiling.
	var stopProf cxprof.StopFunc
	if debugLexer {
		_, stopProf = cxprof.StartProfile(log)
	}

	return log, func() {
		// Dump memory state.
		if err := cxprof.DumpMemProfile(funcName); err != nil {
			log.WithError(err).Error("Failed to dump MEM profile.")
		}

		// Stop Log profiling.
		if stopProf != nil {
			stopProf()
		}

		// Stop CPU profiling.
		if err := stopCPUProf(); err != nil {
			log.WithError(err).Error("Failed to stop CPU profiling.")
		}
	}
}

func compileSources(prog *cxcore.CXProgram, filenames []string, srcs []*os.File) error {
	// Actually parse source code.
	if exitCode := cxlexer.ParseSourceCode(srcs, filenames); exitCode != 0 {
		return fmt.Errorf("cxlexer.ParseSourceCode returned with code %d", exitCode)
	}

	// Set working directory.
	if len(srcs) > 0 {
		cxgo0.PRGRM0.Path = determineWorkDir(srcs[0].Name())
	}

	// Add main function if not exist.
	if _, err := prog.GetFunction(cxcore.MAIN_FUNC, cxcore.MAIN_PKG); err != nil {
		mainPkg := cxcore.MakePackage(cxcore.MAIN_PKG)
		prog.AddPackage(mainPkg)
		mainFn := cxcore.MakeFunction(cxcore.MAIN_FUNC, actions.CurrentFile, actions.LineNo)
		mainPkg.AddFunction(mainFn)
	}

	// Add *init function that initializes all global variables.
	mainPkg, err := prog.GetPackage(cxcore.MAIN_PKG)
	if err != nil {
		return fmt.Errorf("failed to obtain main package: %w", err)
	}
	initFn := cxcore.MakeFunction(cxcore.SYS_INIT_FUNC, actions.CurrentFile, actions.LineNo)
	mainPkg.AddFunction(initFn)
	actions.FunctionDeclaration(initFn, nil, nil, actions.SysInitExprs)
	if _, err := prog.SelectFunction(cxcore.MAIN_FUNC); err != nil {
		return fmt.Errorf("failed to select main package: %w", err)
	}

	// Reset.
	actions.LineNo = 0

	// Check and return.
	if cxcore.FoundCompileErrors {
		return fmt.Errorf("cxcore has compilation error code %d", cxcore.CX_COMPILATION_ERROR)
	}
	return nil
}

func determineWorkDir(filename string) (wkDir string) {
	log := log.WithField("func", "determineWorkDir")
	defer func() {
		log.WithField("work_dir", wkDir).Info()
	}()

	filename = filepath.FromSlash(filename)
	i := strings.LastIndexByte(filename, os.PathSeparator)
	if i == -1 {
		return ""
	}
	return filename[:i]
}
