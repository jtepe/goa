package gencontroller

import (
	"flag"
	"fmt"
	"os"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_main"
	"github.com/goadesign/goa/goagen/utils"
)

//NewGenerator returns an initialized instance of a JavaScript Client Generator
func NewGenerator(options ...Option) *Generator {
	g := &Generator{}

	for _, option := range options {
		option(g)
	}

	return g
}

// Generator is the application code generator.
type Generator struct {
	API       *design.APIDefinition // The API definition
	OutDir    string                // Path to output directory
	DesignPkg string                // Path to design package, only used to mark generated files.
	AppPkg    string                // Name of generated "app" package
	Force     bool                  // Whether to override existing files
	Pkg       string                // Name of the generated package
	Resource  string                // Name of the generated file
	genfiles  []string              // Generated files
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	var (
		outDir, designPkg, appPkg, ver, res, pkg string
		force                                    bool
	)

	set := flag.NewFlagSet("controller", flag.PanicOnError)
	set.StringVar(&outDir, "out", "", "")
	set.StringVar(&designPkg, "design", "", "")
	set.StringVar(&appPkg, "app-pkg", "app", "")
	set.StringVar(&pkg, "pkg", "main", "")
	set.StringVar(&res, "res", "", "")
	set.StringVar(&ver, "version", "", "")
	set.BoolVar(&force, "force", false, "")
	set.Bool("notest", false, "")
	set.Parse(os.Args[1:])

	if err := codegen.CheckVersion(ver); err != nil {
		return nil, err
	}

	appPkg = codegen.Goify(appPkg, false)
	g := &Generator{OutDir: outDir, DesignPkg: designPkg, AppPkg: appPkg, Force: force, API: design.Design, Pkg: pkg, Resource: res}

	return g.Generate()
}

// Generate produces the skeleton controller service factory.
func (g *Generator) Generate() (_ []string, err error) {
	if g.API == nil {
		return nil, fmt.Errorf("missing API definition, make sure design is properly initialized")
	}

	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	if g.AppPkg == "" {
		g.AppPkg = "app"
	}

	codegen.Reserved[g.AppPkg] = true

	err = g.API.IterateResources(func(r *design.ResourceDefinition) error {
		var (
			filename string
			err      error
		)
		if g.Resource == "" || g.Resource == r.Name {
			filename, err = genmain.GenerateControllerFile(g.Force, g.AppPkg, g.OutDir, g.Pkg, r.Name, r)
		}

		if err != nil {
			return err
		}
		g.genfiles = append(g.genfiles, filename)

		return nil
	})

	return g.genfiles, nil
}

// Cleanup removes all the files generated by this generator during the last invokation of Generate.
func (g *Generator) Cleanup() {
	for _, f := range g.genfiles {
		os.Remove(f)
	}
	g.genfiles = nil
}